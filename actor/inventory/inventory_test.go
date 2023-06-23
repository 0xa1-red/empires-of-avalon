package inventory

import (
	"testing"
	"time"

	"github.com/0xa1-red/empires-of-avalon/common"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestBuildingCallback(t *testing.T) {
	g := &Grain{}

	g.buildings = getStartingBuildings()
	g.resources = getStartingResources()

	g.updateLimits()

	g.buildings[common.House].Queue = []Building{
		{
			State:          "inactive",
			WorkersMaximum: 2,
			WorkersCurrent: 0,
			Completion:     time.Now().Add(time.Hour),
		},
	}

	payload := protobuf.TimerFired{
		Timestamp: timestamppb.Now(),
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				KeyBuilding:          structpb.NewStringValue(string(common.House)),
				KeyDisableGenerators: structpb.NewBoolValue(true),
			},
		},
	}

	g.buildingCallback(&payload)

	if expected, actual := 1, len(g.buildings[common.House].Completed); expected != actual {
		t.Fatalf("FAIL: expected amount to be %d, got %d", expected, actual)
	}

	if expected, actual := 0, len(g.buildings[common.House].Queue); expected != actual {
		t.Fatalf("FAIL: expected queue to be %d, got %d", expected, actual)
	}
}

func TestReserveRequest(t *testing.T) {
	grain := &Grain{}

	grain.buildings = getStartingBuildings()
	grain.resources = getStartingResources()

	tests := []struct {
		label          string
		resource       common.ResourceName
		amount         float64
		expectedStatus protobuf.Status
		expectedError  error
	}{
		{
			label:          "success",
			resource:       common.Wood,
			amount:         100,
			expectedStatus: protobuf.Status_OK,
		},
		{
			label:          "insufficient resource error",
			resource:       common.Wood,
			amount:         500,
			expectedStatus: protobuf.Status_Error,
			expectedError:  InsufficientResourceError{Resource: common.Wood},
		},
		{
			label:          "invalid resource error",
			resource:       common.ResourceName("bogus"),
			amount:         1,
			expectedStatus: protobuf.Status_Error,
			expectedError:  InvalidResourceError{Resource: common.ResourceName("bogus")},
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
