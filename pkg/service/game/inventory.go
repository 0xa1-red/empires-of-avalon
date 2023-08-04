package game

import (
	"context"

	gamecluster "github.com/0xa1-red/empires-of-avalon/pkg/cluster"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Describe(ctx context.Context, userID uuid.UUID) (*protobuf.DescribeInventoryResponse, error) {
	inventoryID := GetInventoryID(userID)

	slog.Info("getting inventory grain client", "id", inventoryID.String())
	inventory := protobuf.GetInventoryGrainClient(gamecluster.GetC(), inventoryID.String())

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, &carrier)

	res, err := inventory.Describe(&protobuf.DescribeInventoryRequest{
		TraceID:   carrier.Get("traceparent"),
		Timestamp: timestamppb.Now(),
		GetTimers: false,
	})

	return res, err
}
