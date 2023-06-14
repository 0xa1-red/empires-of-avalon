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
	CallbackBuildings    = "buildings"
	CallbackGenerators   = "generators"
	CallbackTransformers = "transformers"

	KeyBuilding          = "building"
	KeyDisableGenerators = "disable_generators"

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
	subscriptions map[string]*nats.Subscription
	callbacks     map[string]*Callback
}

type Callback struct {
	Name    string
	Subject string
	Method  func(*protobuf.TimerFired)
}

func (g *Grain) Init(ctx cluster.GrainContext) {
	g.ctx = ctx
	g.subscriptions = make(map[string]*nats.Subscription)
	g.callbacks = map[string]*Callback{
		CallbackGenerators: {
			Name:    CallbackGenerators,
			Method:  g.generatorCallback,
			Subject: fmt.Sprintf("%s-resource-callbacks", ctx.Identity()),
		},
		CallbackBuildings: {
			Name:    CallbackBuildings,
			Method:  g.buildingCallback,
			Subject: fmt.Sprintf("%s-building-callbacks", ctx.Identity()),
		},
		CallbackTransformers: {
			Name:    CallbackTransformers,
			Method:  g.transformerCallback,
			Subject: fmt.Sprintf("%s-transform-callbacks", ctx.Identity()),
		},
	}

	// TODO find a better way to only populate if user is new
	g.buildings = getStartingBuildings()
	g.resources = getStartingResources()

	g.updateLimits()

	for _, cb := range g.callbacks {
		if err := g.subscribeToCallback(cb); err != nil {
			slog.Error("failed to subscribe to callback", err, "callback", cb.Name, "subject", cb.Subject)
		}
	}
}

func (g *Grain) updateLimits() {
	for _, resource := range g.resources {
		if err := resource.UpdateCap(g.resources, g.buildings); err != nil {
			slog.Error("failed to calculate resource cap", err)
		}
	}
}

func (g *Grain) Terminate(ctx cluster.GrainContext) {
	defer func() {
		for _, sub := range g.subscriptions {
			sub.Unsubscribe() // nolint
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
		Reply:    g.callbacks[CallbackBuildings].Subject,
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
		finish := ""
		if !meta.Finished.IsZero() && meta.Finished.After(time.Now()) {
			finish = meta.Finished.Format(time.RFC3339)
		}
		buildingValues[string(building)] = structpb.NewStructValue(&structpb.Struct{
			Fields: map[string]*structpb.Value{
				"amount": structpb.NewNumberValue(float64(meta.Amount)),
				"queue":  structpb.NewNumberValue(float64(meta.Queue)),
				"finish": structpb.NewStringValue(finish),
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
		Reply:    g.callbacks[CallbackGenerators].Subject,
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

func (g *Grain) startTransformer(transformer blueprints.Transformer) error {
	slog.Debug("starting transformer", "name", transformer.Name)
	timer := protobuf.GetTimerGrainClient(g.ctx.Cluster(), uuid.New().String())
	res, err := timer.CreateTimer(&protobuf.TimerRequest{
		BuildID:  uuid.New().String(),
		Reply:    g.callbacks[CallbackTransformers].Subject,
		Duration: transformer.TickLength,
		Amount:   -1,
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"cost":   structpb.NewListValue(transformer.CostStructList()),
				"result": structpb.NewListValue(transformer.ResultStructList()),
			},
		},
		Timestamp: timestamppb.Now(),
	})
	if err != nil {
		return nil
	}

	slog.Debug("transform timer response",
		slog.Int("status", int(res.Status)),
		slog.Time("deadline", res.Deadline.AsTime()),
		slog.Time("timestamp", res.Timestamp.AsTime()),
	)

	return nil
}

func (g *Grain) subscribeToCallback(cb *Callback) error {
	subject := cb.Subject
	sub, err := intnats.GetConnection().Subscribe(subject, cb.Method)
	if err != nil {
		return err
	}

	slog.Debug("subscribed to callback subject", "subject", subject)
	g.subscriptions[cb.Name] = sub
	return nil
}

func (g *Grain) buildingCallback(t *protobuf.TimerFired) {
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
	g.buildings[building.Name].Finished = time.Time{}

	for _, cost := range building.Cost {
		if !cost.Permanent {
			g.resources[cost.Resource].Amount += g.resources[cost.Resource].Reserved
		}
		g.resources[cost.Resource].Reserved = 0
	}

	// For testing purposes, we can disable generators if needed
	if disable, ok := payload[KeyDisableGenerators]; ok && disable.(bool) {
		slog.Debug("generators are disabled for building", "building", building.Name)
		return
	}

	for _, gen := range building.Generators {
		if err := g.startGenerator(gen); err != nil {
			slog.Error("failed to start generator", err, "name", gen.Name)
		}
	}

	for _, tf := range building.Transformers {
		if err := g.startTransformer(tf); err != nil {
			slog.Error("failed to start generator", err, "name", tf.Name)
		}
	}
}

func (g *Grain) generatorCallback(t *protobuf.TimerFired) {
	payload := t.Data.AsMap()
	resourceName := payload[KeyResource].(string)
	resource, ok := common.Resources[common.ResourceName(resourceName)]
	if !ok {
		return
	}

	amount := int(payload[KeyAmount].(float64))

	g.resources[resource.Name].Update(amount)
}

func (g *Grain) transformerCallback(t *protobuf.TimerFired) {
	payload := t.Data

	// TODO try to refactor resource allocation
	removeCache := map[string]int{}
	addCache := map[string]int{}
	rollback := func() {
		for name, amount := range removeCache {
			g.resources[common.ResourceName(name)].Amount += amount
		}
	}

	for _, cost := range payload.Fields["cost"].GetListValue().Values {
		c := cost.GetStructValue()
		resource := c.Fields["resource"].GetStringValue()
		g.resources[common.ResourceName(resource)].mx.Lock()
		current := g.resources[common.ResourceName(resource)].Amount
		needed := int(c.Fields["amount"].GetNumberValue())

		if current < needed {
			slog.Debug("insufficient resource", "name", resource, "current", current, "needed", needed)
			rollback()
			g.resources[common.ResourceName(resource)].mx.Unlock()
			return
		}

		g.resources[common.ResourceName(resource)].Amount -= needed
		removeCache[resource] = needed
		g.resources[common.ResourceName(resource)].mx.Unlock()
	}

	for _, result := range payload.Fields["result"].GetListValue().Values {
		r := result.GetStructValue()
		resource := r.Fields["resource"].GetStringValue()
		g.resources[common.ResourceName(resource)].mx.Lock()
		added := int(r.Fields["amount"].GetNumberValue())

		g.resources[common.ResourceName(resource)].Amount += added
		addCache[resource] = added
		g.resources[common.ResourceName(resource)].mx.Unlock()
	}

	slog.Info("transformer callback fired", "removed", removeCache, "added", addCache)
	// spew.Dump(t.Data.AsMap())
}

func (g *Grain) Reserve(req *protobuf.ReserveRequest, ctx cluster.GrainContext) (*protobuf.ReserveResponse, error) {
	resources := req.Resources.AsMap()

	var err error
	cache := make(map[common.ResourceName]int)
	for resource, amount := range resources {
		amt := int(amount.(float64))
		resourceName := common.ResourceName(resource)
		if current, ok := g.resources[resourceName]; ok {
			current.mx.Lock()
			if current.Amount >= amt {
				current.Amount -= amt
				current.Reserved += amt
				cache[resourceName] = amt
			} else {
				err = InsufficientResourceError{Resource: resourceName}
			}
			current.mx.Unlock()
		} else {
			err = InvalidResourceError{Resource: resourceName}
		}
	}

	if err != nil {
		for resource, amount := range cache {
			g.resources[resource].Amount += amount
			g.resources[resource].Reserved -= amount
		}

		return &protobuf.ReserveResponse{
			Timestamp: timestamppb.Now(),
			Status:    protobuf.Status_Error,
			Error:     err.Error(),
		}, nil
	}

	return &protobuf.ReserveResponse{
		Timestamp: timestamppb.Now(),
		Status:    protobuf.Status_OK,
	}, nil
}
