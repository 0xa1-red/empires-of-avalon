package inventory

import (
	"fmt"
	"sync"
	"time"

	"github.com/0xa1-red/empires-of-avalon/common"
	"github.com/0xa1-red/empires-of-avalon/persistence"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	intnats "github.com/0xa1-red/empires-of-avalon/transport/nats"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BuildingRegister struct {
	mx *sync.Mutex

	Amount   int
	Queue    int
	Finished time.Time
}

type Grain struct {
	ctx cluster.GrainContext

	buildings    map[common.Building]*BuildingRegister
	replySubject string
	subscription *nats.Subscription
}

func (g *Grain) Init(ctx cluster.GrainContext) {
	g.ctx = ctx
	label := fmt.Sprintf("%s-subject", ctx.Identity())
	g.replySubject = uuid.NewSHA1(uuid.NameSpaceOID, []byte(label)).String()
	g.buildings = make(map[common.Building]*BuildingRegister)

	cb := func(t *protobuf.TimerFired) {
		payload := t.Data.AsMap()
		buildingName := payload["building"].(string)
		building, ok := common.Buildings[common.BuildingName(buildingName)]
		if !ok {
			slog.Error("failed to complete building", fmt.Errorf("Invalid building name: %s", buildingName))
			return
		}
		slog.Debug("finished building", "building", building.Name)
		g.buildings[building].Amount += 1
		g.buildings[building].Queue -= 1
	}

	sub, err := intnats.GetConnection().Subscribe(g.replySubject, cb)
	if err != nil {
		slog.Error("failed to subscribe to reply subject", err)
		return
	}

	g.subscription = sub
}
func (g *Grain) Terminate(ctx cluster.GrainContext) {
	defer g.subscription.Unsubscribe()
	if len(g.buildings) == 0 {
		return
	}

	if n, err := persistence.Get().Persist(g); err != nil {
		slog.Error("failed to persist grain", err, "kind", g.Kind(), "identity", ctx.Identity())
	} else {
		slog.Debug("grain successfully persisted", "kind", g.Kind(), "identity", ctx.Identity(), "written", n)
	}

}

func (g *Grain) ReceiveDefault(ctx cluster.GrainContext) {}

func (g *Grain) Start(req *protobuf.StartRequest, ctx cluster.GrainContext) (*protobuf.StartResponse, error) {
	b, ok := common.Buildings[common.BuildingName(req.Name)]
	if !ok {
		return &protobuf.StartResponse{
			Status:    protobuf.Status_Error,
			Error:     fmt.Sprintf("Invalid building name: %s", req.Name),
			Timestamp: timestamppb.Now(),
		}, nil
	}
	if _, ok := g.buildings[b]; !ok {
		g.buildings[b] = &BuildingRegister{
			mx: &sync.Mutex{},
		}
	}
	if g.buildings[b].Queue > 0 {
		return &protobuf.StartResponse{
			Status:    protobuf.Status_Error,
			Error:     fmt.Sprintf("Building is already in progress: %s", req.Name),
			Timestamp: timestamppb.Now(),
		}, nil
	}

	slog.Info("requested building", "name", string(b.Name))

	timer := protobuf.GetTimerGrainClient(g.ctx.Cluster(), uuid.New().String())
	res, err := timer.CreateTimer(&protobuf.TimerRequest{
		BuildID:  uuid.New().String(),
		Reply:    g.replySubject,
		Duration: b.BuildTime,
		Amount:   req.Amount,
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"building": structpb.NewStringValue(string(b.Name)),
			},
		},
		Timestamp: timestamppb.Now(),
	})
	if err != nil {
		return &protobuf.StartResponse{
			Status:    protobuf.Status_Error,
			Error:     err.Error(),
			Timestamp: timestamppb.Now(),
		}, nil
	}

	slog.Info("timer response",
		slog.Int("status", int(res.Status)),
		slog.Time("deadline", res.Deadline.AsTime()),
		slog.Time("timestamp", res.Timestamp.AsTime()),
	)

	g.buildings[b].Queue += int(req.Amount)
	d, _ := time.ParseDuration(b.BuildTime)
	start := time.Now()
	for r := req.Amount; r > 0; r-- {
		start = start.Add(d)
	}
	g.buildings[b].Finished = start

	return &protobuf.StartResponse{
		Status:    protobuf.Status_OK,
		Timestamp: timestamppb.Now(),
	}, nil
}

func (g *Grain) Describe(_ *protobuf.DescribeInventoryRequest, ctx cluster.GrainContext) (*protobuf.DescribeInventoryResponse, error) {
	values := make(map[string]*structpb.Value)

	for building, meta := range g.buildings {
		values[string(building.Name)] = structpb.NewStructValue(&structpb.Struct{
			Fields: map[string]*structpb.Value{
				"amount": structpb.NewNumberValue(float64(meta.Amount)),
				"queue":  structpb.NewNumberValue(float64(meta.Queue)),
				"finish": structpb.NewStringValue(meta.Finished.Format(time.RFC3339)),
			},
		})
	}

	inventory := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"buildings": structpb.NewStructValue(&structpb.Struct{
				Fields: values,
			}),
		},
	}

	return &protobuf.DescribeInventoryResponse{
		Inventory: inventory,
		Timestamp: timestamppb.Now(),
	}, nil
}
