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

	g.buildings = g.getStartingBuildings()
	g.resources = g.getStartingResources()

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

	if expected, actual := 1, len(g.buildings[blueprintID].Completed); expected != actual {
		t.Fatalf("FAIL: expected amount to be %d, got %d", expected, actual)
	}

	if expected, actual := 0, len(g.buildings[blueprintID].Queue); expected != actual {
		t.Fatalf("FAIL: expected queue to be %d, got %d", expected, actual)
	}
}

func TestReserveRequest(t *testing.T) {
	grain := &Grain{}

	if err := setupRegistry(); err != nil {
		t.Fatalf("Fail: %v", err)
	}

	grain.buildings = grain.getStartingBuildings()
	grain.resources = grain.getStartingResources()

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
			if res == nil {
				t.Fatalf("Fail: expected response, got nil")
			}
			if actual, expected := res.Status, tt.expectedStatus; actual != expected {
				t.Fatalf("FAIL: expected status to be %s, got %s", expected, actual)
			}
			if tt.expectedStatus == protobuf.Status_Error {
				if actual, expected := res.Error, tt.expectedError.Error(); actual != expected {
					t.Fatalf("FAIL: expected error be %s, got %s", expected, actual)
				}
			}
		}

		t.Run(tt.label, tf)
	}
}
