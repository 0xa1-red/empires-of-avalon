package inventory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/0xa1-red/empires-of-avalon/actor"
	"github.com/0xa1-red/empires-of-avalon/instrumentation/traces"
	"github.com/0xa1-red/empires-of-avalon/persistence"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/game"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	intnats "github.com/0xa1-red/empires-of-avalon/transport/nats"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	CallbackBuildings    = "buildings"
	CallbackGenerators   = "generators"
	CallbackTransformers = "transformers"
	CallbackTimerStopped = "timer-stopped"

	SubjectTimerStatus = "timer-status"

	KeyBuilding          = "building"
	KeyId                = "id"
	KeyDisableGenerators = "disable_generators"

	KeyResource = "resource"
	KeyAmount   = "amount"

	StateQueued   = "queued"
	StateInactive = "inactive"
	StateActive   = "active"
)

type ReservedResource struct {
	Name      string
	Amount    int
	Permanent bool
}

type BuildingRegister struct {
	mx *sync.Mutex

	BlueprintID uuid.UUID
	Name        blueprints.BuildingName
	Completed   map[uuid.UUID]Building
	Queue       map[uuid.UUID]Building
}

type TimerRegister struct {
	mx *sync.Mutex

	Transformers []uuid.UUID
	Generators   []uuid.UUID
}

func NewTimerRegister() *TimerRegister {
	return &TimerRegister{
		mx: &sync.Mutex{},

		Transformers: make([]uuid.UUID, 0),
		Generators:   make([]uuid.UUID, 0),
	}
}

type Grain struct {
	ctx cluster.GrainContext

	buildings       map[uuid.UUID]*BuildingRegister
	resources       map[blueprints.ResourceName]*ResourceRegister
	subscriptions   map[string]*nats.Subscription
	callbacks       map[string]*Callback
	heartbeatTicker *time.Ticker
	timers          map[uuid.UUID]struct{}
}

type Callback struct {
	Name    string
	Subject string
	Method  func(*protobuf.TimerFired)
}

func (g *Grain) Init(ctx cluster.GrainContext) {
	g.ctx = ctx
	g.subscriptions = make(map[string]*nats.Subscription)
	g.timers = make(map[uuid.UUID]struct{})
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

	g.buildings = g.getStartingBuildings()
	g.resources = g.getStartingResources()

	g.updateLimits()

	for _, cb := range g.callbacks {
		if err := g.subscribeToCallback(cb); err != nil {
			slog.Error("failed to subscribe to callback", err,
				"callback", cb.Name,
				"subject", cb.Subject,
			)
		}
	}

	for blueprintID, register := range g.buildings {
		bp, err := registry.GetBuilding(blueprintID)
		if err != nil {
			slog.Error("failed to retrieve building blueprint", err, "id", blueprintID)
		}

		for buildingID := range register.Completed {
			g.startBuildingGenerators(buildingID, bp)
			g.startBuildingTransformers(buildingID, bp)
		}
	}

	if err := g.subscribeToTimerStopped(); err != nil {
		slog.Error("failed to subscribe to callback", err,
			"callback", CallbackTimerStopped,
			"subject", SubjectTimerStatus,
		)
	}

	if err := actor.SendUpdate(&protobuf.GrainUpdate{ // nolint:exhaustruct
		UpdateKind: protobuf.UpdateKind_Register,
		GrainKind:  protobuf.GrainKind_InventoryGrain,
		Timestamp:  timestamppb.Now(),
		Identity:   g.ctx.Self().String(),
	}); err != nil {
		slog.Warn("failed to send register update to admin actor", err)
	}

	g.heartbeatTicker = time.NewTicker(30 * time.Second)
	go func() {
		for curTime := range g.heartbeatTicker.C {
			if err := actor.SendUpdate(&protobuf.GrainUpdate{ // nolint:exhaustruct
				UpdateKind: protobuf.UpdateKind_Heartbeat,
				GrainKind:  protobuf.GrainKind_InventoryGrain,
				Timestamp:  timestamppb.New(curTime),
				Identity:   g.ctx.Self().String(),
			}); err != nil {
				slog.Warn("failed to send register update to admin actor", err)
			}
		}
	}()
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

	g.heartbeatTicker.Stop()

	if n, err := persistence.Get().Persist(g); err != nil {
		slog.Error("failed to persist grain", err, "kind", g.Kind(), "identity", ctx.Identity())
	} else {
		slog.Debug("grain successfully persisted", "kind", g.Kind(), "identity", ctx.Identity(), "written", n)
	}
}

func (g *Grain) ReceiveDefault(ctx cluster.GrainContext) {}

func (g *Grain) StartBuilding(req *protobuf.StartBuildingRequest, ctx cluster.GrainContext) (*protobuf.StartBuildingResponse, error) {
	carrier := propagation.MapCarrier{}
	carrier.Set("traceparent", req.TraceID)
	pctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

	sctx, span := traces.Start(pctx, "actor/inventory/start")
	defer span.End()

	blueprint, err := registry.GetBuilding(game.GetBuildingID(req.Name))
	if err != nil {
		return &protobuf.StartBuildingResponse{
			Status:    protobuf.Status_Error,
			Error:     fmt.Sprintf("Invalid building name: %s", req.Name),
			Timestamp: timestamppb.Now(),
		}, nil
	}

	if _, ok := g.buildings[blueprint.ID]; !ok {
		g.buildings[blueprint.ID] = &BuildingRegister{
			mx:          &sync.Mutex{},
			Name:        blueprint.Name,
			Completed:   make(map[uuid.UUID]Building),
			Queue:       make(map[uuid.UUID]Building),
			BlueprintID: blueprint.ID,
		}
	}

	queue := 0
	for _, b := range g.buildings {
		queue += len(b.Queue)
	}

	if queue > 0 {
		return &protobuf.StartBuildingResponse{
			Status:    protobuf.Status_Error,
			Error:     "All building slots are occupied",
			Timestamp: timestamppb.Now(),
		}, nil
	}

	insufficient := make([]string, 0)

	for _, cost := range blueprint.Cost {
		if g.resources[cost.Resource].Amount < cost.Amount {
			insufficient = append(insufficient, string(cost.Resource))
		}
	}

	if len(insufficient) > 0 {
		return &protobuf.StartBuildingResponse{
			Status:    protobuf.Status_Error,
			Error:     fmt.Sprintf("Insufficient resources: %s", strings.Join(insufficient, ", ")),
			Timestamp: timestamppb.Now(),
		}, nil
	}

	buildingID := uuid.New()
	slog.Info("requested building", "name", string(blueprint.Name), "id", buildingID.String())

	otel.GetTextMapPropagator().Inject(sctx, &carrier)

	timerID := uuid.New()
	timer := protobuf.GetTimerGrainClient(g.ctx.Cluster(), timerID.String())
	res, err := timer.CreateTimer(&protobuf.TimerRequest{
		TimerID:     timerID.String(),
		TraceID:     carrier.Get("traceparent"),
		Kind:        protobuf.TimerKind_Building,
		Reply:       g.callbacks[CallbackBuildings].Subject,
		InventoryID: g.ctx.Identity(),
		Duration:    blueprint.BuildTime,
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				KeyId:       structpb.NewStringValue(buildingID.String()),
				KeyBuilding: structpb.NewStringValue(string(blueprint.Name)),
				KeyAmount:   structpb.NewNumberValue(1),
			},
		},
		Timestamp: timestamppb.Now(),
	})

	if err != nil {
		return &protobuf.StartBuildingResponse{
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

	g.timers[timerID] = struct{}{}

	reserved := make([]ReservedResource, 0)
	for _, resource := range blueprint.Cost {
		reserved = append(reserved, ReservedResource{
			Name:      string(resource.Resource),
			Amount:    resource.Amount,
			Permanent: resource.Permanent,
		})
	}

	build := Building{
		ID:                buildingID,
		BlueprintID:       blueprint.ID,
		Name:              blueprint.Name,
		State:             protobuf.BuildingState_BuildingStateQueued,
		WorkersMaximum:    blueprint.WorkersMaximum,
		WorkersCurrent:    0,
		Completion:        res.Deadline.AsTime(),
		ReservedResources: reserved,
		Timers:            NewTimerRegister(),
	}

	g.buildings[blueprint.ID].Queue[buildingID] = build

	return &protobuf.StartBuildingResponse{
		Status:    protobuf.Status_OK,
		Timestamp: timestamppb.Now(),
		Error:     "",
	}, nil
}

func (g *Grain) Describe(req *protobuf.DescribeInventoryRequest, ctx cluster.GrainContext) (*protobuf.DescribeInventoryResponse, error) {
	carrier := propagation.MapCarrier{}
	carrier.Set("traceparent", req.TraceID)
	pctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

	nctx, span := traces.Start(pctx, "actor/inventory/describe")
	defer span.End()

	buildingValues := g.describeBuildings(nctx)
	resourceValues := g.describeResources(nctx)

	fields := map[string]*structpb.Value{
		"buildings": structpb.NewStructValue(&structpb.Struct{
			Fields: buildingValues,
		}),
		"resources": structpb.NewStructValue(&structpb.Struct{
			Fields: resourceValues,
		}),
	}

	if req.GetTimers {
		t := make(map[string]interface{})
		for k, v := range g.timers {
			t[k.String()] = v
		}

		timers, err := structpb.NewStruct(t)
		if err == nil {
			fields["timers"] = structpb.NewStructValue(timers)
		} else {
			slog.Warn("failed to collect timers for describe response", err)
		}
	}

	inventory := &structpb.Struct{
		Fields: fields,
	}

	return &protobuf.DescribeInventoryResponse{
		Inventory: inventory,
		Timestamp: timestamppb.Now(),
	}, nil
}

func (g *Grain) getStartingBuildings() map[uuid.UUID]*BuildingRegister {
	registers := make(map[uuid.UUID]*BuildingRegister)

	now := time.Now()

	for blueprintID, blueprint := range registry.GetBuildings() {
		completed := make(map[uuid.UUID]Building, 0)

		for i := 0; i < blueprint.InitialAmount; i++ {
			id := uuid.New()
			completed[id] = Building{
				ID:                id,
				BlueprintID:       blueprintID,
				Name:              blueprint.Name,
				State:             protobuf.BuildingState_BuildingStateActive,
				WorkersMaximum:    blueprint.WorkersMaximum,
				WorkersCurrent:    0,
				Completion:        now,
				ReservedResources: make([]ReservedResource, 0),
				Timers:            NewTimerRegister(),
			}
		}

		registers[blueprint.ID] = &BuildingRegister{
			mx:          &sync.Mutex{},
			Name:        blueprint.Name,
			Completed:   completed,
			Queue:       make(map[uuid.UUID]Building),
			BlueprintID: blueprintID,
		}
	}

	return registers
}

func (g *Grain) startGenerator(generator blueprints.Generator) (uuid.UUID, error) {
	slog.Debug("starting generator", "name", generator.Name)

	timerID := uuid.New()

	timer := protobuf.GetTimerGrainClient(g.ctx.Cluster(), timerID.String())
	res, err := timer.CreateTimer(&protobuf.TimerRequest{
		TimerID:     timerID.String(),
		TraceID:     "",
		Kind:        protobuf.TimerKind_Generator,
		Reply:       g.callbacks[CallbackGenerators].Subject,
		Duration:    generator.TickLength,
		InventoryID: g.ctx.Identity(),
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"resource": structpb.NewStringValue(string(generator.Name)),
				"amount":   structpb.NewNumberValue(float64(generator.Amount)),
			},
		},
		Timestamp: timestamppb.Now(),
	})

	if err != nil {
		return uuid.Nil, nil
	}

	slog.Debug("resource timer response",
		slog.Int("status", int(res.Status)),
		slog.Time("deadline", res.Deadline.AsTime()),
		slog.Time("timestamp", res.Timestamp.AsTime()),
	)

	return timerID, nil
}

func (g *Grain) startTransformer(transformer blueprints.Transformer) (uuid.UUID, error) {
	slog.Debug("starting transformer", "name", transformer.Name)

	timerID := uuid.New()

	timer := protobuf.GetTimerGrainClient(g.ctx.Cluster(), timerID.String())
	res, err := timer.CreateTimer(&protobuf.TimerRequest{
		TimerID:     timerID.String(),
		TraceID:     "",
		Kind:        protobuf.TimerKind_Transformer,
		Reply:       g.callbacks[CallbackTransformers].Subject,
		Duration:    transformer.TickLength,
		InventoryID: g.ctx.Identity(),
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"cost":   structpb.NewListValue(transformer.CostStructList()),
				"result": structpb.NewListValue(transformer.ResultStructList()),
			},
		},
		Timestamp: timestamppb.Now(),
	})

	if err != nil {
		slog.Debug("transform timer returned error",
			slog.Int("status", int(res.Status)),
			slog.Time("timestamp", res.Timestamp.AsTime()),
		)

		return uuid.Nil, err
	}

	slog.Debug("transform timer response",
		slog.Int("status", int(res.Status)),
		slog.Time("deadline", res.Deadline.AsTime()),
		slog.Time("timestamp", res.Timestamp.AsTime()),
	)

	return timerID, nil
}

func (g *Grain) subscribeToCallback(cb *Callback) error {
	subject := cb.Subject

	transport, err := intnats.GetConnection()
	if err != nil {
		return err
	}

	sub, err := transport.Subscribe(subject, cb.Method)

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
	buildingIDStr := payload[KeyId].(string)

	buildingID, err := uuid.Parse(buildingIDStr)
	if err != nil {
		slog.Warn("failed to parse building ID", "raw", buildingIDStr)
		return
	}

	blueprint, err := registry.GetBuilding(game.GetBuildingID(buildingName))
	if err != nil {
		slog.Warn("failed to retrieve blueprint from registry", "name", buildingName)
		return
	}

	g.buildings[blueprint.ID].mx.Lock()
	defer g.buildings[blueprint.ID].mx.Unlock()

	slog.Debug("finished building", "building", blueprint.Name)
	b := g.buildings[blueprint.ID].Queue[buildingID]
	b.State = protobuf.BuildingState_BuildingStateActive
	b.Completion = time.Now()
	g.buildings[blueprint.ID].Completed[buildingID] = b

	delete(g.buildings[blueprint.ID].Queue, buildingID)

	for _, cost := range blueprint.Cost {
		if !cost.Permanent {
			g.resources[cost.Resource].Amount += g.resources[cost.Resource].Reserved
		}

		g.resources[cost.Resource].Reserved = 0
	}

	// For testing purposes, we can disable generators if needed
	if disable, ok := payload[KeyDisableGenerators]; ok && disable.(bool) {
		slog.Debug("generators are disabled for building", "building", blueprint.Name)
		return
	}

	g.startBuildingGenerators(buildingID, blueprint)
	g.startBuildingTransformers(buildingID, blueprint)
}

func (g *Grain) generatorCallback(t *protobuf.TimerFired) {
	payload := t.Data.AsMap()
	resourceName := payload[KeyResource].(string)

	resource, err := registry.GetResource(resourceName)
	if err != nil {
		return
	}

	amount := int(payload[KeyAmount].(float64))

	g.resources[resource.Name].Update(amount)
}

func (g *Grain) transformerCallback(t *protobuf.TimerFired) {
	payload := t.Data

	reserveCache := map[string]int{}
	addCache := map[string]int{}

	for _, result := range payload.Fields["result"].GetListValue().Values {
		r := result.GetStructValue()
		resource := r.Fields["resource"].GetStringValue()
		g.resources[blueprints.ResourceName(resource)].mx.Lock()
		added := int(r.Fields["amount"].GetNumberValue())

		g.resources[blueprints.ResourceName(resource)].Amount += added

		addCache[resource] = added
		g.resources[blueprints.ResourceName(resource)].mx.Unlock()
	}

	for _, cost := range payload.Fields["cost"].GetListValue().Values {
		r := cost.GetStructValue()

		resource := r.Fields["resource"].GetStringValue()
		g.resources[blueprints.ResourceName(resource)].mx.Lock()
		reserved := int(r.Fields["amount"].GetNumberValue())

		if r.Fields["temporary"].GetBoolValue() {
			g.resources[blueprints.ResourceName(resource)].Amount += reserved
			addCache[resource] += reserved
		}

		g.resources[blueprints.ResourceName(resource)].Reserved -= reserved
		reserveCache[resource] = reserved
		g.resources[blueprints.ResourceName(resource)].mx.Unlock()
	}

	slog.Info("transformer callback fired", "removed", reserveCache, "added", addCache)
}

func (g *Grain) timerStoppedCallback(t *protobuf.TimerStopped) {
	id, err := uuid.Parse(t.TimerID)
	if err != nil {
		slog.Error("failed to parse timer ID", err, "timer_id", t.TimerID, "timestamp", t.Timestamp.AsTime().Format(time.RFC1123))
		return
	}

	delete(g.timers, id)
	slog.Debug("removed timer from register", "timer_id", t.TimerID, "timestamp", t.Timestamp.AsTime().Format(time.RFC1123))
}

func (g *Grain) Reserve(req *protobuf.ReserveRequest, ctx cluster.GrainContext) (*protobuf.ReserveResponse, error) {
	carrier := propagation.MapCarrier{}
	carrier.Set("traceparent", req.TraceID)
	pctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

	_, span := traces.Start(pctx, "actor/inventory/reserve")
	defer span.End()

	resources := req.Resources.AsMap()

	var err error

	cache := make(map[blueprints.ResourceName]int)

	for resource, amount := range resources {
		amt := int(amount.(float64))
		resourceName := blueprints.ResourceName(resource)

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
			slog.Info("reserved resources", "resource", resource, "amount", amt)
		} else {
			err = InvalidResourceError{Resource: resourceName}
		}
	}

	if err != nil {
		for resource, amount := range cache {
			g.resources[resource].Amount += amount
			g.resources[resource].Reserved -= amount
			slog.Info("rolling back reserved resources", "resource", resource, "amount", amount)
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
		Error:     "",
	}, nil
}

func (g *Grain) startBuildingGenerators(buildingID uuid.UUID, b *blueprints.Building) {
	timers := make([]uuid.UUID, 0)

	for _, gen := range b.Generates {
		if timerID, err := g.startGenerator(gen); err != nil {
			slog.Error("failed to start generator", err, "name", gen.Name)
		} else {
			timers = append(timers, timerID)
		}
	}

	buildingRegister, ok := g.buildings[b.ID]
	if !ok {
		slog.Warn("building register not found", "inventory_id", g.Identity(), "blueprint_id", b.ID)
		return
	}

	completedBuilding, ok := buildingRegister.Completed[buildingID]
	if !ok {
		slog.Warn("building not found", "inventory_id", g.Identity(), "blueprint_id", b.ID, "building_id", buildingID)
		return
	}

	completedBuilding.Timers.Generators = timers
}

func (g *Grain) startBuildingTransformers(buildingID uuid.UUID, b *blueprints.Building) {
	timers := make([]uuid.UUID, 0)

	for _, tr := range b.Transforms {
		if timerID, err := g.startTransformer(tr); err != nil {
			slog.Error("failed to start transformer", err, "name", tr.Name)
		} else {
			timers = append(timers, timerID)
		}
	}

	g.buildings[b.ID].Completed[buildingID].Timers.Transformers = timers
}

func (g *Grain) subscribeToTimerStopped() error {
	subject := "timer-status"

	transport, err := intnats.GetConnection()
	if err != nil {
		return err
	}

	sub, err := transport.Subscribe(subject, g.timerStoppedCallback)

	if err != nil {
		return err
	}

	slog.Debug("subscribed to callback subject", "subject", subject)

	g.subscriptions[CallbackTimerStopped] = sub

	return nil
}

func (g *Grain) describeBuildings(ctx context.Context) map[string]*structpb.Value {
	_, span := traces.Start(ctx, "actor/inventory/describe_buildings")
	defer span.End()

	buildingValues := make(map[string]*structpb.Value)

	for _, meta := range g.buildings {
		completedList := make([]*structpb.Value, 0)

		for _, state := range meta.Completed {
			completedList = append(completedList, state.Describe())
		}

		queuedList := make([]*structpb.Value, 0)

		for _, state := range meta.Queue {
			queuedList = append(queuedList, state.Describe())
		}

		completed := structpb.NewListValue(&structpb.ListValue{
			Values: completedList,
		})
		queued := structpb.NewListValue(&structpb.ListValue{
			Values: queuedList,
		})

		buildingValues[meta.Name.String()] = structpb.NewStructValue(&structpb.Struct{
			Fields: map[string]*structpb.Value{
				"completed": completed,
				"queued":    queued,
			},
		})
	}

	return buildingValues
}

func (g *Grain) describeResources(ctx context.Context) map[string]*structpb.Value {
	_, span := traces.Start(ctx, "actor/inventory/describe_resources")
	defer span.End()

	resourceValues := make(map[string]*structpb.Value)

	for resource, meta := range g.resources {
		resourceValues[string(resource)] = structpb.NewStructValue(&structpb.Struct{
			Fields: map[string]*structpb.Value{
				"amount":   structpb.NewNumberValue(float64(meta.Amount)),
				"reserved": structpb.NewNumberValue(float64(meta.Reserved)),
				"cap":      structpb.NewNumberValue(float64(meta.Cap)),
			},
		})
	}

	return resourceValues
}
