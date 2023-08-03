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
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/exp/slog"
)

var AdminID = uuid.NewSHA1(uuid.NameSpaceOID, []byte("avalon.admin"))
var AdminSubject = fmt.Sprintf("admin-updates-%s", AdminID.String())

type actor struct {
	Identity    string
	PID         *pactor.PID
	LastSeen    time.Time
	Kind        protobuf.GrainKind
	Tolerations int
}

type registry struct {
	Inventories map[string]actor
	Timers      map[string]actor
}

func (g *Grain) add(a actor) {
	a.PID = PIDFromIdentity(a.Identity)

	switch a.Kind {
	case protobuf.GrainKind_InventoryGrain:
		g.registry.Inventories[a.Identity] = a
	case protobuf.GrainKind_TimerGrain:
		g.registry.Timers[a.Identity] = a
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
	switch a.Kind {
	case protobuf.GrainKind_InventoryGrain:
		delete(g.registry.Inventories, a.Identity)
	case protobuf.GrainKind_TimerGrain:
		delete(g.registry.Timers, a.Identity)
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
		g.registry.Inventories[a.Identity] = a
	case protobuf.GrainKind_TimerGrain:
		g.registry.Timers[a.Identity] = a
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
			for _, i := range g.registry.Inventories {
				if i.LastSeen.Before(time.Now().Add(-1 * time.Minute)) {
					i.Tolerations += 1

					slog.Warn(
						"grain has not been reporting for the last minute",
						"kind", i.Kind.String(),
						"identity", i.Identity,
						"last_seen", i.LastSeen.Format(time.RFC1123),
						"tolerations", i.Tolerations,
					)

					if i.Tolerations >= 3 {
						slog.Warn("grain failed to report more than three times",
							"kind", i.Kind.String(),
							"identity", i.Identity,
							"last_seen", i.LastSeen.Format(time.RFC1123),
							"tolerations", i.Tolerations,
						)
					}
				}
			}

			for _, i := range g.registry.Timers {
				if i.LastSeen.Before(time.Now().Add(-1 * time.Minute)) {
					i.Tolerations += 1

					slog.Warn(
						"grain has not been reporting for the last minute",
						"kind", i.Kind.String(),
						"identity", i.Identity,
						"last_seen", i.LastSeen.Format(time.RFC1123),
						"tolerations", i.Tolerations,
					)
				}
			}
		}
	}()

	return nil, nil
}

func (g *Grain) messageCallback(t *protobuf.GrainUpdate) {
	a := actor{
		Identity:    t.Identity,
		LastSeen:    t.Timestamp.AsTime(),
		Kind:        t.GrainKind,
		Tolerations: 0,
		PID:         nil,
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
