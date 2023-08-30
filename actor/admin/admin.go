package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/0xa1-red/empires-of-avalon/instrumentation/metrics"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	intnats "github.com/0xa1-red/empires-of-avalon/transport/nats"
	pactor "github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var AdminID = uuid.NewSHA1(uuid.NameSpaceOID, []byte("avalon.admin"))
var AdminSubject = fmt.Sprintf("admin-updates-%s", AdminID.String())

type actor struct {
	Identity    string
	PID         *ActorPID
	LastSeen    time.Time
	Kind        protobuf.GrainKind
	Tolerations int
	TimerKind   protobuf.TimerKind
	Context     map[string]interface{}
}

type ActorPID struct {
	*pactor.PID

	GrainID uuid.UUID
}

func (a ActorPID) GetGrainID() uuid.UUID {
	return a.GrainID
}

func (a actor) AsMap() map[string]interface{} {
	m := map[string]interface{}{
		"identity":    a.Identity,
		"grain_id":    a.PID.GrainID.String(),
		"pid":         a.PID.String(),
		"last_seen":   a.LastSeen.Format(time.RFC1123),
		"kind":        a.Kind.String(),
		"tolerations": a.Tolerations,
		"context":     a.Context,
	}

	return m
}

type registry struct {
	Inventories map[string]actor
	Timers      map[string]actor
}

func (g *Grain) add(a actor) {
	pid, err := PIDFromIdentity(a.Identity)
	if err != nil {
		slog.Error("failed to get PID from identity", err, "identity", a.Identity)
	}

	a.PID = pid

	switch a.Kind {
	case protobuf.GrainKind_InventoryGrain:
		g.registry.Inventories[a.PID.GrainID.String()] = a
	case protobuf.GrainKind_TimerGrain:
		g.registry.Timers[a.PID.GrainID.String()] = a
	default:
		slog.Warn("unknown grain kind", "kind", a.Kind.Number())
		return
	}

	g.activeActors.Add(context.Background(), 1, metric.WithAttributes(
		attribute.String("kind", a.Kind.String()),
	))
	slog.Debug("added grain to registry", "identity", a.Identity, "kind", a.Kind.String())
}

func (g *Grain) remove(a actor) {
	id := a.PID.GrainID.String()

	switch a.Kind {
	case protobuf.GrainKind_InventoryGrain:
		if _, ok := g.registry.Inventories[id]; !ok {
			slog.Warn("grain doesn't exist in registry", "kind", a.Kind.String(), "id", id)
			return
		}

		delete(g.registry.Inventories, id)
	case protobuf.GrainKind_TimerGrain:
		if _, ok := g.registry.Timers[id]; !ok {
			slog.Warn("grain doesn't exist in registry", "kind", a.Kind.String(), "id", id)
			return
		}

		slog.Debug("removing timer grain", "id", id)

		delete(g.registry.Timers, id)

		spew.Dump(g.registry.Timers)
	default:
		slog.Warn("unknown grain kind", "kind", a.Kind.Number())
		return
	}

	g.activeActors.Add(context.Background(), -1, metric.WithAttributes(
		attribute.String("kind", a.Kind.String()),
	))
	slog.Debug("removed grain from registry", "identity", a.Identity, "kind", a.Kind.String())
}

func (g *Grain) heartbeat(a actor) {
	switch a.Kind {
	case protobuf.GrainKind_InventoryGrain:
		g.registry.Inventories[a.PID.GrainID.String()] = a
	case protobuf.GrainKind_TimerGrain:
		g.registry.Timers[a.PID.GrainID.String()] = a
	default:
		slog.Warn("unknown grain kind", "kind", a.Kind.Number())
		return
	}

	slog.Debug("updated grain in registry", "identity", a.Identity, "kind", a.Kind.String())
}

type Grain struct {
	ctx cluster.GrainContext

	registry     *registry
	subscription *nats.Subscription
	activeActors metric.Int64UpDownCounter
	cleanupTimer *time.Ticker
}

func (g *Grain) Init(ctx cluster.GrainContext) {
	g.ctx = ctx

	var err error
	if g.activeActors, err = metrics.Meter().Int64UpDownCounter("actors_active"); err != nil {
		slog.Warn("failed to register actors_active instrument", "error", err)
	}

	g.activeActors.Add(context.Background(), 1, metric.WithAttributes(
		attribute.String("kind", protobuf.GrainKind_AdminGrain.String()),
	))
}

func (g *Grain) Terminate(ctx cluster.GrainContext) {
	if g.cleanupTimer != nil {
		g.cleanupTimer.Stop()
	}
}

func (g *Grain) ReceiveDefault(ctx cluster.GrainContext) {}

func (g *Grain) Start(_ *protobuf.Empty, ctx cluster.GrainContext) (*protobuf.Empty, error) {
	slog.Info("spawning admin actor", "identity", ctx.Identity())

	g.registry = &registry{
		Inventories: make(map[string]actor),
		Timers:      make(map[string]actor),
	}

	transport, err := intnats.GetConnection()
	if err != nil {
		return nil, err
	}

	sub, err := transport.Subscribe(AdminSubject, g.messageCallback)
	if err != nil {
		return nil, err
	}

	slog.Debug("subscribed to admin callback subject", "subject", AdminSubject)

	g.subscription = sub

	g.cleanupTimer = time.NewTicker(30 * time.Second)
	go func() {
		for range g.cleanupTimer.C {
			for i := range g.registry.Inventories {
				actor := g.registry.Inventories[i]
				checkHeartbeat(&actor)
			}

			for i := range g.registry.Timers {
				actor := g.registry.Timers[i]
				checkHeartbeat(&actor)
			}
		}
	}()

	return nil, nil
}

func checkHeartbeat(a *actor) {
	if a.LastSeen.Before(time.Now().Add(-1 * time.Minute)) {
		a.Tolerations += 1

		slog.Warn(
			"grain has not been reporting for the last minute",
			"kind", a.Kind.String(),
			"identity", a.Identity,
			"last_seen", a.LastSeen.Format(time.RFC1123),
			"tolerations", a.Tolerations,
		)

		if a.Tolerations >= 3 {
			slog.Warn("grain failed to report more than three times",
				"kind", a.Kind.String(),
				"identity", a.Identity,
				"last_seen", a.LastSeen.Format(time.RFC1123),
				"tolerations", a.Tolerations,
			)
		}
	}
}

func (g *Grain) messageCallback(t *protobuf.GrainUpdate) {
	pid, err := PIDFromIdentity(t.Identity)
	if err != nil {
		slog.Error("failed to get PID from identity", err, "identity", t.Identity)
	}

	a := actor{ // nolint
		Identity:    t.Identity,
		LastSeen:    t.Timestamp.AsTime(),
		Kind:        t.GrainKind,
		Tolerations: 0,
		PID:         pid,
		Context:     t.Context.AsMap(),
	}

	switch t.UpdateKind {
	case protobuf.UpdateKind_Register:
		g.add(a)
	case protobuf.UpdateKind_Deregister:
		g.remove(a)
	case protobuf.UpdateKind_Heartbeat:
		g.heartbeat(a)
	default:
		slog.Warn("unknown update kind", "kind", t.UpdateKind.Number())
		return
	}
}

func (g *Grain) Describe(req *protobuf.DescribeAdminRequest, ctx cluster.GrainContext) (*protobuf.DescribeAdminResponse, error) {
	inventoryRegistry := make(map[string]interface{})
	for k, v := range g.registry.Inventories {
		inventoryRegistry[k] = v.AsMap()
	}

	timerRegistry := make(map[string]interface{})
	for k, v := range g.registry.Timers {
		timerRegistry[k] = v.AsMap()
	}

	registry := map[string]interface{}{
		"inventories": inventoryRegistry,
		"timers":      timerRegistry,
	}

	admin := map[string]interface{}{
		"registry": registry,
	}

	adminStruct, err := structpb.NewStruct(admin)
	if err != nil {
		return &protobuf.DescribeAdminResponse{
			Admin:     nil,
			Timestamp: timestamppb.Now(),
			Status:    protobuf.Status_Error,
			Error:     err.Error(),
		}, nil
	}

	return &protobuf.DescribeAdminResponse{
		Admin:     adminStruct,
		Timestamp: timestamppb.Now(),
		Status:    protobuf.Status_OK,
		Error:     "",
	}, nil
}
