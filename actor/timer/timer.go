package timer

import (
	"fmt"
	"time"

	"github.com/0xa1-red/empires-of-avalon/persistence"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/0xa1-red/empires-of-avalon/transport/nats"
	"github.com/asynkron/protoactor-go/cluster"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Timer struct {
	Kind        protobuf.TimerKind
	InventoryID string
	Reply       string
	Amount      int64
	Start       time.Time
	Interval    time.Duration
	Data        map[string]interface{}
}

type Grain struct {
	ctx   cluster.GrainContext
	timer *Timer
}

func (g *Grain) Init(ctx cluster.GrainContext) {
	g.ctx = ctx
}
func (g *Grain) Terminate(ctx cluster.GrainContext) {
	if g.timer.Amount > 0 {
		if n, err := persistence.Get().Persist(g); err != nil {
			slog.Error("failed to persist grain", err, "kind", g.Kind(), "identity", ctx.Identity())
		} else {
			slog.Debug("grain successfully persisted", "kind", g.Kind(), "identity", ctx.Identity(), "written", n)
		}
	}
}
func (g *Grain) ReceiveDefault(ctx cluster.GrainContext) {}

func (g *Grain) CreateTimer(req *protobuf.TimerRequest, ctx cluster.GrainContext) (*protobuf.TimerResponse, error) {
	start := time.Now()
	d, err := time.ParseDuration(req.Duration)
	if err != nil {
		return &protobuf.TimerResponse{
			Status:    protobuf.Status_Error,
			Error:     err.Error(),
			Timestamp: timestamppb.New(start),
		}, nil
	}

	g.timer = &Timer{
		Kind:        req.Kind,
		Reply:       req.Reply,
		Start:       start,
		Interval:    d,
		InventoryID: req.InventoryID,
		Data:        req.Data.AsMap(),
	}

	slog.Info("starting timer", "trace_id", req.TraceID, "interval", d.String())
	timerFn := g.startBuildingTimer
	switch req.Kind {
	case protobuf.TimerKind_Generator:
		timerFn = g.startGenerateTimer
	case protobuf.TimerKind_Transformer:
		timerFn = g.startTransformTimer
	}

	go timerFn()

	deadline := start
	deadline = deadline.Add(d)

	return &protobuf.TimerResponse{
		Status:    protobuf.Status_OK,
		Deadline:  timestamppb.New(deadline),
		Timestamp: timestamppb.Now(),
	}, nil
}

func (g *Grain) startBuildingTimer() {
	now := time.Now()
	conn := nats.GetConnection()
	d, err := structpb.NewValue(g.timer.Data)
	if err != nil {
		slog.Error("failed to start timer", err)
	}

	if g.timer != nil {
		nextTrigger := g.timer.Start.Add(g.timer.Interval)
		if nextTrigger.Before(now) {
			if err := conn.Publish(g.timer.Reply, &protobuf.TimerFired{
				Timestamp: timestamppb.New(now),
				Data:      d.GetStructValue(),
			}); err != nil {
				slog.Error("failed to send TimerFired message", err)
			}

			g.timer = nil
			return
		}
	}

	t := time.NewTimer(g.timer.Interval)

	for curTime := range t.C {
		slog.Debug("timer fired", "kind", g.timer.Kind.String(), "reply", g.timer.Reply, "inventory", g.timer.InventoryID)
		if err := conn.Publish(g.timer.Reply, &protobuf.TimerFired{
			Timestamp: timestamppb.New(curTime),
			Data:      d.GetStructValue(),
		}); err != nil {
			slog.Error("failed to send TimerFired message", err)
		}

		if g.timer.Amount == 0 {
			t.Stop()
		}
	}
}

func (g *Grain) startGenerateTimer() {
	now := time.Now()
	conn := nats.GetConnection()
	d, err := structpb.NewValue(g.timer.Data)
	if err != nil {
		slog.Error("failed to start timer", err)
	}

	for {
		nextTrigger := g.timer.Start.Add(g.timer.Interval)
		if nextTrigger.Before(now) {
			if err := conn.Publish(g.timer.Reply, &protobuf.TimerFired{
				Timestamp: timestamppb.New(now),
				Data:      d.GetStructValue(),
			}); err != nil {
				slog.Error("failed to send TimerFired message", err)
			}
			g.timer.Start = nextTrigger
		} else {
			break
		}
	}

	t := time.NewTicker(g.timer.Interval)

	for curTime := range t.C {
		slog.Debug("timer fired", "kind", g.timer.Kind.String(), "reply", g.timer.Reply, "inventory", g.timer.InventoryID)
		if err := conn.Publish(g.timer.Reply, &protobuf.TimerFired{
			Timestamp: timestamppb.New(curTime),
			Data:      d.GetStructValue(),
		}); err != nil {
			slog.Error("failed to send TimerFired message", err)
		}
	}
}

func (g *Grain) startTransformTimer() {
	now := time.Now()
	conn := nats.GetConnection()
	d, err := structpb.NewValue(g.timer.Data)
	if err != nil {
		slog.Error("failed to start timer", err)
	}

	for {
		nextTrigger := g.timer.Start.Add(g.timer.Interval)
		if nextTrigger.Before(now) {
			if err := conn.Publish(g.timer.Reply, &protobuf.TimerFired{
				Timestamp: timestamppb.New(now),
				Data:      d.GetStructValue(),
			}); err != nil {
				slog.Error("failed to send TimerFired message", err)
			}
			g.timer.Start = nextTrigger
		} else {
			break
		}
	}

	for {
		send := true
		if err := g.reserveResources(); err != nil {
			send = false
		}

		t := time.NewTimer(g.timer.Interval)

		curTime := <-t.C
		if send {
			slog.Debug("timer fired", "kind", g.timer.Kind.String(), "reply", g.timer.Reply, "inventory", g.timer.InventoryID)
			if err := conn.Publish(g.timer.Reply, &protobuf.TimerFired{
				Timestamp: timestamppb.New(curTime),
				Data:      d.GetStructValue(),
			}); err != nil {
				slog.Error("failed to send TimerFired message", err)
			}
		} else {
			slog.Error("timer skipped, because of reserve error", err)
		}
	}
}

func (g *Grain) reserveResources() error {
	resources, err := getResourcesFromTimer(g.timer)
	if err != nil {
		return fmt.Errorf("failed to get resources: %w", err)
	}

	slog.Debug("reserving resources", "inventory", g.timer.InventoryID, "resources", resources)

	r, err := structpb.NewValue(resources)
	if err != nil {
		return err
	}
	ig := protobuf.GetInventoryGrainClient(g.ctx.Cluster(), g.timer.InventoryID)
	msg := protobuf.ReserveRequest{
		Resources: r.GetStructValue(),
		Timestamp: timestamppb.Now(),
	}

	res, err := ig.Reserve(&msg)
	if err != nil {
		return err
	}

	if res.Status == protobuf.Status_Error {
		return fmt.Errorf("%s", res.Error)
	}

	return nil
}

func getResourcesFromTimer(t *Timer) (map[string]interface{}, error) {
	resources := make(map[string]interface{})
	costs, ok := t.Data["cost"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid cost list value")
	}

	for _, cost := range costs {
		c, ok := cost.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid cost value")
		}

		var cc string
		var amt float64

		if cc, ok = c["resource"].(string); !ok {
			return nil, fmt.Errorf("invalid resource value")
		}
		if amt, ok = c["amount"].(float64); !ok {
			return nil, fmt.Errorf("invalid amount value")
		}
		resources[cc] = int(amt)
	}

	return resources, nil
}
