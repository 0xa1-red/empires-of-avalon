package inventory

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/game"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupRegistry() error {
	absPath, _ := filepath.Abs("./testdata") // nolint:errcheck
	slog.Debug("intializing registry", "blueprint_path", absPath)

	if buildings, err := registry.ReadYaml[*blueprints.Building](absPath); err != nil {
		slog.Error("failed to read in building blueprints", err)
		return err
	} else {
		for _, building := range buildings {
			if err := registry.Push(building); err != nil {
				slog.Warn("failed to push building blueprint to local registry", "err", err)
			}
		}
	}

	if resources, err := registry.ReadYaml[*blueprints.Resource](absPath); err != nil {
		slog.Error("failed to read in resource blueprints", err)
		return err
	} else {
		for _, resource := range resources {
			if err := registry.Push(resource); err != nil {
				slog.Warn("failed to push resource blueprint to local registry", "err", err)
			}
		}
	}

	return nil
}

func TestBuildingCallback(t *testing.T) {
	g := &Grain{}

	if err := setupRegistry(); err != nil {
		t.Fatalf("Fail: %v", err)
	}

	var assetError error

	g.buildings, assetError = g.getStartingBuildings()
	assert.NoError(t, assetError)

	g.resources, assetError = g.getStartingResources()
	assert.NoError(t, assetError)

	g.updateLimits()

	blueprintID := game.GetBuildingID(blueprints.House.String())
	buildingID := uuid.New()

	g.buildings[blueprintID].Queue = map[uuid.UUID]Building{
		buildingID: {
			ID:             buildingID,
			BlueprintID:    blueprintID,
			State:          protobuf.BuildingState_BuildingStateInactive,
			WorkersMaximum: 2,
			WorkersCurrent: 0,
			Completion:     time.Now().Add(time.Hour),
		},
	}

	payload := protobuf.TimerFired{
		Timestamp: timestamppb.Now(),
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				KeyBuilding:          structpb.NewStringValue(string(blueprints.House)),
				KeyDisableGenerators: structpb.NewBoolValue(true),
				KeyId:                structpb.NewStringValue(buildingID.String()),
			},
		},
	}

	g.buildingCallback(&payload)

	assert.Equal(t, 1, len(g.buildings[blueprintID].Completed))
	assert.Equal(t, 0, len(g.buildings[blueprintID].Queue))
}

func TestReserveRequest(t *testing.T) {
	grain := &Grain{}

	if err := setupRegistry(); err != nil {
		t.Fatalf("Fail: %v", err)
	}

	var assetError error

	grain.buildings, assetError = grain.getStartingBuildings()
	assert.NoError(t, assetError)

	grain.resources, assetError = grain.getStartingResources()
	assert.NoError(t, assetError)

	tests := []struct {
		label          string
		resource       blueprints.ResourceName
		amount         float64
		expectedStatus protobuf.Status
		expectedError  error
	}{
		{
			label:          "success",
			resource:       blueprints.Wood,
			amount:         100,
			expectedStatus: protobuf.Status_OK,
		},
		{
			label:          "insufficient resource error",
			resource:       blueprints.Wood,
			amount:         500,
			expectedStatus: protobuf.Status_Error,
			expectedError:  InsufficientResourceError{Resource: blueprints.Wood},
		},
		{
			label:          "invalid resource error",
			resource:       blueprints.ResourceName("bogus"),
			amount:         1,
			expectedStatus: protobuf.Status_Error,
			expectedError:  InvalidResourceError{Resource: blueprints.ResourceName("bogus")},
		},
	}

	for _, tt := range tests {
		tf := func(t *testing.T) {
			msg := protobuf.ReserveRequest{
				Resources: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						string(tt.resource): structpb.NewNumberValue(tt.amount),
					},
				},
			}

			res, _ := grain.Reserve(&msg, nil)
			assert.NotNil(t, res)

			assert.Equal(t, tt.expectedStatus, res.Status)

			if tt.expectedStatus == protobuf.Status_Error {
				assert.Equal(t, tt.expectedError.Error(), res.Error)
			}
		}

		t.Run(tt.label, tf)
	}
}
