package inventory

import (
	"time"

	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/structpb"
)

type Building struct {
	ID                uuid.UUID
	BlueprintID       uuid.UUID
	Name              blueprints.BuildingName
	State             protobuf.BuildingState
	WorkersMaximum    int
	WorkersCurrent    int
	Completion        time.Time
	ReservedResources []ReservedResource
	Timers            *TimerRegister
}

func (b Building) Describe() *structpb.Value {
	finish := ""
	if !b.Completion.IsZero() {
		finish = b.Completion.Format(time.RFC3339)
	}

	var fields map[string]*structpb.Value

	if b.State == protobuf.BuildingState_BuildingStateQueued {
		fields = map[string]*structpb.Value{
			"state":      structpb.NewStringValue(b.State.String()),
			"completion": structpb.NewStringValue(finish),
		}
	} else {
		generators := make([]interface{}, 0)
		for _, gen := range b.Timers.Generators {
			generators = append(generators, gen.String())
		}

		genpb, err := structpb.NewList(generators)
		if err != nil {
			slog.Warn("failed to marshal generator list")
		}

		transformers := make([]interface{}, 0)
		for _, trans := range b.Timers.Transformers {
			transformers = append(transformers, trans.String())
		}

		transpb, err := structpb.NewList(transformers)
		if err != nil {
			slog.Warn("failed to marshal generator list")
		}

		fields = map[string]*structpb.Value{
			"id":              structpb.NewStringValue(b.ID.String()),
			"state":           structpb.NewStringValue(b.State.String()),
			"workers_max":     structpb.NewNumberValue(float64(b.WorkersMaximum)),
			"workers_current": structpb.NewNumberValue(float64(b.WorkersCurrent)),
			"completion":      structpb.NewStringValue(finish),
			"generators":      structpb.NewListValue(genpb),
			"transformers":    structpb.NewListValue(transpb),
		}
	}

	return structpb.NewStructValue(&structpb.Struct{
		Fields: fields,
	})
}
