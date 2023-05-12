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

	g.buildings[common.House].Queue = 1
	g.buildings[common.House].Finished = time.Now().Add(time.Hour)

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

	if expected, actual := 1, g.buildings[common.House].Amount; expected != actual {
		t.Fatalf("FAIL: expected amount to be %d, got %d", expected, actual)
	}

	if expected, actual := 0, g.buildings[common.House].Queue; expected != actual {
		t.Fatalf("FAIL: expected queue to be %d, got %d", expected, actual)
	}

	if expected, actual := true, g.buildings[common.House].Finished.IsZero(); expected != actual {
		t.Fatalf("FAIL: expected finished to be zero, got %s", g.buildings[common.House].Finished.Format(time.RFC3339))
	}
}
