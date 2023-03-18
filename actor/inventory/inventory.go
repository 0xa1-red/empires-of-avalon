package inventory

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/0xa1-red/empires-of-avalon/blueprints"
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

const (
	CallbackBuildings = "buildings"
	CallbackResources = "resources"

	KeyBuilding = "building"

	KeyResource = "resource"
	KeyAmount   = "amount"
)

type BuildingRegister struct {
	mx *sync.Mutex

	Name     common.BuildingName
	Amount   int
	Queue    int
	Finished time.Time
}

type Grain struct {
	ctx cluster.GrainContext

	buildings     map[common.BuildingName]*BuildingRegister
	resources     map[common.ResourceName]*ResourceRegister
	replySubjects map[string]string
	subscriptions map[string]*nats.Subscription
}

func (g *Grain) Init(ctx cluster.GrainContext) {
	g.ctx = ctx
	g.subscriptions = make(map[string]*nats.Subscription)
	g.replySubjects = map[string]string{
		CallbackBuildings: fmt.Sprintf("%s-building-callbacks", ctx.Identity()),
		CallbackResources: fmt.Sprintf("%s-resource-callbacks", ctx.Identity()),
	}

	// TODO find a better way to only populate if user is new
	g.buildings = getStartingBuildings()
	g.resources = getStartingResources()

	g.updateLimits()

	if err := g.subscribeToBuildingCallbacks(); err != nil {
		slog.Error("failed to subscribe to building callbacks", err, "subject", g.replySubjects[CallbackBuildings])
		return
	}

	if err := g.subscribeToResourceCallbacks(); err != nil {
		slog.Error("failed to subscribe to resource callbacks", err, "subject", g.replySubjects[CallbackResources])
		return
	}
}

func (g *Grain) updateLimits() {
	newLimits := make(map[common.ResourceName]int)
	for _, b := range g.buildings {
		for resource, limit := range common.Buildings[b.Name].Limits {
			if _, ok := newLimits[resource]; !ok {
				newLimits[resource] = 0
			}

			newLimits[resource] += b.Amount * limit
		}
	}

	for resource, limit := range newLimits {
		slog.Debug("setting new cap for resource", "name", resource, "cap", limit)
		g.resources[resource].Cap = limit
		g.resources[resource].Update(0)
	}
}

func (g *Grain) Terminate(ctx cluster.GrainContext) {
	defer func() {
		for _, sub := range g.subscriptions {
			sub.Unsubscribe()
		}
	}()
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

	queue := 0
	for _, b := range g.buildings {
		queue += b.Queue
	}
	if queue > 0 {
		return &protobuf.StartResponse{
			Status:    protobuf.Status_Error,
			Error:     "All building slots are occupied",
			Timestamp: timestamppb.Now(),
		}, nil
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

	slog.Info("requested building", logFields...)

	timer := protobuf.GetTimerGrainClient(g.ctx.Cluster(), uuid.New().String())
	res, err := timer.CreateTimer(&protobuf.TimerRequest{
		BuildID:  uuid.New().String(),
		Reply:    g.replySubjects[CallbackBuildings],
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

	slog.Debug("building timer response",
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
				"cap":      structpb.NewNumberValue(float64(meta.Cap)),
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

func (g *Grain) startGenerator(generator blueprints.Generator) error {
	slog.Debug("starting generator", "name", generator.Name)
	timer := protobuf.GetTimerGrainClient(g.ctx.Cluster(), uuid.New().String())
	res, err := timer.CreateTimer(&protobuf.TimerRequest{
		BuildID:  uuid.New().String(),
		Reply:    g.replySubjects[CallbackResources],
		Duration: generator.TickLength,
		Amount:   -1,
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"resource": structpb.NewStringValue(string(generator.Name)),
				"amount":   structpb.NewNumberValue(float64(generator.Amount)),
			},
		},
		Timestamp: timestamppb.Now(),
	})
	if err != nil {
		return nil
	}

	slog.Debug("resource timer response",
		slog.Int("status", int(res.Status)),
		slog.Time("deadline", res.Deadline.AsTime()),
		slog.Time("timestamp", res.Timestamp.AsTime()),
	)

	return nil
}

func (g *Grain) subscribeToResourceCallbacks() error {
	cb := func(t *protobuf.TimerFired) {
		payload := t.Data.AsMap()
		resourceName := payload[KeyResource].(string)
		resource, ok := common.Resources[common.ResourceName(resourceName)]
		if !ok {
			return
		}

		amount := int(payload[KeyAmount].(float64))

		g.resources[resource.Name].Update(amount)
	}

	sub, err := intnats.GetConnection().Subscribe(g.replySubjects[CallbackResources], cb)
	if err != nil {
		return err
	}

	g.subscriptions[CallbackResources] = sub
	return nil
}

func (g *Grain) subscribeToBuildingCallbacks() error {
	cb := func(t *protobuf.TimerFired) {
		defer g.updateLimits()
		payload := t.Data.AsMap()
		buildingName := payload[KeyBuilding].(string)
		building, ok := common.Buildings[common.BuildingName(buildingName)]
		if !ok {
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

		for _, gen := range building.Generators {
			if err := g.startGenerator(gen); err != nil {
				slog.Error("failed to start generator", err, "name", gen.Name)
			}
		}
	}

	sub, err := intnats.GetConnection().Subscribe(g.replySubjects[CallbackBuildings], cb)
	if err != nil {
		return err
	}

	g.subscriptions[CallbackBuildings] = sub
	return nil
}
