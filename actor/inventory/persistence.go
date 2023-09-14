package inventory

import (
	"sync"

	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (g *Grain) Encode() ([]byte, error) {
	snapshot := &protobuf.InventorySnapshot{ // nolint:exhaustruct
		Buildings: buildingRegistryPb(g.buildings),
		Resources: resourceRegistryPb(g.resources),
		Timers:    timersPb(g.timers),
	}

	return proto.Marshal(snapshot)
}

func buildingRegistryPb(buildings map[uuid.UUID]*BuildingRegister) []*protobuf.InventoryBuildingRegistry {
	collection := make([]*protobuf.InventoryBuildingRegistry, 0)

	for _, reg := range buildings {
		completed := make([]*protobuf.InventoryBuilding, 0)
		for _, bb := range reg.Completed {
			completed = append(completed, buildingPb(bb))
		}

		b := &protobuf.InventoryBuildingRegistry{ // nolint:exhaustruct
			BlueprintID: reg.BlueprintID.String(),
			Name:        reg.Name.String(),
			Completed:   completed,
		}
		b.BlueprintID = reg.BlueprintID.String()
		b.Name = reg.Name.String()
		collection = append(collection, b)
	}

	return collection
}

func buildingPb(bb Building) *protobuf.InventoryBuilding {
	copy := &protobuf.InventoryBuilding{
		ID:                bb.ID.String(),
		BlueprintID:       bb.BlueprintID.String(),
		Name:              bb.Name.String(),
		State:             bb.State,
		WorkersMax:        int32(bb.WorkersMaximum),
		WorkersCurr:       int32(bb.WorkersMaximum),
		Completed:         timestamppb.New(bb.Completion),
		ReservedResources: reservedResourcesPb(bb.ReservedResources),
		Generators:        buildingTimersPb(bb.Timers.Generators),
		Transformers:      buildingTimersPb(bb.Timers.Transformers),
	}

	return copy
}

func reservedResourcesPb(rr []ReservedResource) []*protobuf.ReservedResource {
	pb := make([]*protobuf.ReservedResource, 0)
	for _, resource := range rr {
		pb = append(pb, &protobuf.ReservedResource{
			Name:      resource.Name,
			Amount:    int64(resource.Amount),
			Permanent: resource.Permanent,
		})
	}

	return pb
}

func buildingTimersPb(timers []uuid.UUID) []*protobuf.BuildingTimer {
	pb := make([]*protobuf.BuildingTimer, 0)
	for _, t := range timers {
		pb = append(pb, &protobuf.BuildingTimer{
			ID: t.String(),
		})
	}

	return pb
}

func resourceRegistryPb(resources map[blueprints.ResourceName]*ResourceRegister) []*protobuf.InventoryResourceRegistry {
	pb := make([]*protobuf.InventoryResourceRegistry, 0)
	for _, res := range resources {
		pb = append(pb, &protobuf.InventoryResourceRegistry{
			Name:       res.Name.String(),
			CapFormula: res.CapFormula,
			Cap:        int64(res.Cap),
			Amount:     int64(res.Amount),
			Reserved:   int64(res.Reserved),
		})
	}

	return pb
}

func timersPb(timers map[uuid.UUID]struct{}) []*protobuf.InventoryTimer {
	pb := make([]*protobuf.InventoryTimer, 0)

	for id := range timers {
		pb = append(pb, &protobuf.InventoryTimer{
			ID: id.String(),
		})
	}

	return pb
}

func (g *Grain) restore(snapshot *protobuf.InventorySnapshot) error {
	buildings := make(map[uuid.UUID]*BuildingRegister)

	for _, reg := range snapshot.Buildings {
		blueprintID, err := uuid.Parse(reg.BlueprintID)
		if err != nil {
			return err
		}

		register := &BuildingRegister{ // nolint:exhaustruct
			BlueprintID: blueprintID,
			Name:        blueprints.BuildingName(reg.Name),
			Completed:   make(map[uuid.UUID]Building),
			Queue:       make(map[uuid.UUID]Building),
		}

		for _, building := range reg.Completed {
			b, err := restoreBuilding(building, blueprintID)
			if err != nil {
				return err
			}

			register.Completed[b.ID] = b
		}

		buildings[blueprintID] = register
	}

	g.buildings = buildings

	return nil
}

func restoreResources(pb []*protobuf.ReservedResource) []ReservedResource {
	resources := make([]ReservedResource, 0)

	for _, res := range pb {
		resource := ReservedResource{
			Name:      res.Name,
			Amount:    int(res.Amount),
			Permanent: res.Permanent,
		}
		resources = append(resources, resource)
	}

	return resources
}

func restoreTimers(pb []*protobuf.BuildingTimer) ([]uuid.UUID, error) {
	timers := make([]uuid.UUID, 0)

	for _, timer := range pb {
		id, err := uuid.Parse(timer.ID)
		if err != nil {
			return nil, err
		}

		timers = append(timers, id)
	}

	return timers, nil
}

func restoreBuilding(building *protobuf.InventoryBuilding, blueprintID uuid.UUID) (Building, error) {
	b := Building{} // nolint:exhaustruct

	buildingID, err := uuid.Parse(building.ID)
	if err != nil {
		return b, err
	}

	transformers, err := restoreTimers(building.Transformers)
	if err != nil {
		return b, err
	}

	generators, err := restoreTimers(building.Generators)
	if err != nil {
		return b, err
	}

	b = Building{
		ID:                buildingID,
		BlueprintID:       blueprintID,
		Name:              blueprints.BuildingName(building.Name),
		State:             building.State,
		WorkersMaximum:    int(building.WorkersMax),
		WorkersCurrent:    int(building.WorkersCurr),
		Completion:        building.Completed.AsTime(),
		ReservedResources: restoreResources(building.ReservedResources),
		Timers: &TimerRegister{
			mx: &sync.Mutex{},

			Transformers: transformers,
			Generators:   generators,
		},
	}

	return b, nil
}
