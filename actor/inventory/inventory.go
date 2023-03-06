package inventory

import (
	"fmt"
	"strings"
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

	Name     common.BuildingName
	Amount   int
	Queue    int
	Finished time.Time
}

type ResourceRegister struct {
	mx *sync.Mutex

	Name     common.ResourceName
	Amount   int
	Reserved int
}

type Grain struct {
	ctx cluster.GrainContext

	buildings    map[common.BuildingName]*BuildingRegister
	resources    map[common.ResourceName]*ResourceRegister
	replySubject string
	subscription *nats.Subscription
}

func (g *Grain) Init(ctx cluster.GrainContext) {
	g.ctx = ctx
	label := fmt.Sprintf("%s-subject", ctx.Identity())
	g.replySubject = uuid.NewSHA1(uuid.NameSpaceOID, []byte(label)).String()

	// TODO find a better way to only populate if user is new
	g.buildings = getStartingBuildings()
	g.resources = getStartingResources()

	cb := func(t *protobuf.TimerFired) {
		payload := t.Data.AsMap()
		buildingName := payload["building"].(string)
		building, ok := common.Buildings[common.BuildingName(buildingName)]
		if !ok {
			slog.Error("failed to complete building", fmt.Errorf("Invalid building name: %s", buildingName))
			return
		}
		slog.Debug("finished building", "building", building.Name)
		g.buildings[building.Name].Amount += 1
		g.buildings[building.Name].Queue -= 1

		for _, cost := range building.Cost {
			if !cost.Permanent {
				g.resources[cost.Resource].Amount += g.resources[cost.Resource].Reserved
			}
			g.resources[cost.Resource].Reserved = 0
		}
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
	if _, ok := g.buildings[b.Name]; !ok {
		g.buildings[b.Name] = &BuildingRegister{
			mx:   &sync.Mutex{},
			Name: b.Name,
		}
	}

	insufficient := make([]string, 0)
	for _, cost := range b.Cost {
		if g.resources[cost.Resource].Amount < cost.Amount {
			insufficient = append(insufficient, string(cost.Resource))
		}
	}

	if len(insufficient) > 0 {
		return &protobuf.StartResponse{
			Status:    protobuf.Status_Error,
			Error:     fmt.Sprintf("Insufficient resources: %s", strings.Join(insufficient, ", ")),
			Timestamp: timestamppb.Now(),
		}, nil
	}

	logFields := []any{"name", string(b.Name)}
	for _, cost := range b.Cost {
		g.resources[cost.Resource].Amount -= cost.Amount
		g.resources[cost.Resource].Reserved = cost.Amount
		logFields = append(logFields, string(cost.Resource), cost.Amount)
	}

	if g.buildings[b.Name].Queue > 0 {
		return &protobuf.StartResponse{
			Status:    protobuf.Status_Error,
			Error:     fmt.Sprintf("Building is already in progress: %s", req.Name),
			Timestamp: timestamppb.Now(),
		}, nil
	}

	slog.Info("requested building", logFields...)

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

	g.buildings[b.Name].Queue += int(req.Amount)
	d, _ := time.ParseDuration(b.BuildTime)
	start := time.Now()
	for r := req.Amount; r > 0; r-- {
		start = start.Add(d)
	}
	g.buildings[b.Name].Finished = start

	return &protobuf.StartResponse{
		Status:    protobuf.Status_OK,
		Timestamp: timestamppb.Now(),
	}, nil
}

func (g *Grain) Describe(_ *protobuf.DescribeInventoryRequest, ctx cluster.GrainContext) (*protobuf.DescribeInventoryResponse, error) {
	buildingValues := make(map[string]*structpb.Value)
	resourceValues := make(map[string]*structpb.Value)

	for building, meta := range g.buildings {
		buildingValues[string(building)] = structpb.NewStructValue(&structpb.Struct{
			Fields: map[string]*structpb.Value{
				"amount": structpb.NewNumberValue(float64(meta.Amount)),
				"queue":  structpb.NewNumberValue(float64(meta.Queue)),
				"finish": structpb.NewStringValue(meta.Finished.Format(time.RFC3339)),
			},
		})
	}

	for resource, meta := range g.resources {
		resourceValues[string(resource)] = structpb.NewStructValue(&structpb.Struct{
			Fields: map[string]*structpb.Value{
				"amount":   structpb.NewNumberValue(float64(meta.Amount)),
				"reserved": structpb.NewNumberValue(float64(meta.Reserved)),
			},
		})
	}

	inventory := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"buildings": structpb.NewStructValue(&structpb.Struct{
				Fields: buildingValues,
			}),
			"resources": structpb.NewStructValue(&structpb.Struct{
				Fields: resourceValues,
			}),
		},
	}

	return &protobuf.DescribeInventoryResponse{
		Inventory: inventory,
		Timestamp: timestamppb.Now(),
	}, nil
}

func getStartingResources() map[common.ResourceName]*ResourceRegister {
	registers := make(map[common.ResourceName]*ResourceRegister)

	for name := range common.Resources {
		registers[name] = &ResourceRegister{
			mx:       &sync.Mutex{},
			Name:     name,
			Amount:   0,
			Reserved: 0,
		}
	}

	registers[common.Population] = newResourceRegister(common.Population, 5)
	registers[common.Wood] = newResourceRegister(common.Wood, 100)

	return registers
}

func getStartingBuildings() map[common.BuildingName]*BuildingRegister {
	registers := make(map[common.BuildingName]*BuildingRegister)

	for name := range common.Buildings {
		registers[name] = &BuildingRegister{
			mx:     &sync.Mutex{},
			Name:   name,
			Amount: 0,
			Queue:  0,
		}
	}

	return registers
}

func newResourceRegister(name common.ResourceName, amount int) *ResourceRegister {
	return &ResourceRegister{
		mx:       &sync.Mutex{},
		Name:     name,
		Amount:   amount,
		Reserved: 0,
	}
}
