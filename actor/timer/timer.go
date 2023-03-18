package timer

import (
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
	Reply    string
	Amount   int64
	Start    time.Time
	Interval time.Duration
	Data     map[string]interface{}
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
		Reply:    req.Reply,
		Amount:   req.Amount,
		Start:    start,
		Interval: d,
		Data:     req.Data.AsMap(),
	}

	slog.Info("starting timer", "trace_id", req.TraceID, "amount", req.Amount, "interval", d.String())
	go g.startTimer()

	deadline := start
	for i := int64(0); i < req.Amount; i++ {
		deadline = deadline.Add(d)
	}

	return &protobuf.TimerResponse{
		Status:    protobuf.Status_OK,
		Deadline:  timestamppb.New(deadline),
		Timestamp: timestamppb.Now(),
	}, nil
}

func (g *Grain) startTimer() {
	now := time.Now()
	conn := nats.GetConnection()
	d, err := structpb.NewValue(g.timer.Data)
	if err != nil {
		slog.Error("failed to start timer", err)
	}

	for {
		if g.timer == nil || g.timer.Amount == 0 {
			break
		}
		nextTrigger := g.timer.Start.Add(g.timer.Interval)
		if nextTrigger.Before(now) {
			if err := conn.Publish(g.timer.Reply, &protobuf.TimerFired{
				Timestamp: timestamppb.New(now),
				Data:      d.GetStructValue(),
			}); err != nil {
				slog.Error("failed to send TimerFired message", err)
			}
			if g.timer.Amount != -1 {
				g.timer.Amount--
			}
		} else {
			break
		}
	}

	if g.timer.Amount == 0 {
		g.timer = nil
		return
	}

	t := time.NewTicker(g.timer.Interval)

	for curTime := range t.C {
		if g.timer.Amount != -1 {
			g.timer.Amount--
		}
		slog.Debug("timer fired", "reply", g.timer.Reply)
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
