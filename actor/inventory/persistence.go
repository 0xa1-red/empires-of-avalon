// nolint:unused
package inventory

import (
	"sync"

	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/persistence"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ persistence.Persistable = (*Grain)(nil)
var _ persistence.Restorable = (*Grain)(nil)

func (g *Grain) Encode() ([]byte, error) {
	snapshot := &protobuf.InventorySnapshot{
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

		b := &protobuf.InventoryBuildingRegistry{
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

func (g *Grain) Restore(raw []byte) error {
	snapshot := &protobuf.InventorySnapshot{}
	if err := proto.Unmarshal(raw, snapshot); err != nil {
		return err
	}

	buildings, err := restoreBuildingRegisters(snapshot.Buildings)
	if err != nil {
		return err
	}

	resources, err := restoreResourceRegisters(snapshot.Resources)
	if err != nil {
		return err
	}

	timers, err := restoreTimers(snapshot.Timers)
	if err != nil {
		return err
	}

	g.buildings = buildings
	g.resources = resources
	g.timers = timers

	return nil
}

func restoreBuildingRegisters(pb []*protobuf.InventoryBuildingRegistry) (map[uuid.UUID]*BuildingRegister, error) {
	buildings := make(map[uuid.UUID]*BuildingRegister)

	for _, reg := range pb {
		blueprintID, err := uuid.Parse(reg.BlueprintID)
		if err != nil {
			return nil, err
		}

		register := &BuildingRegister{
			BlueprintID: blueprintID,
			Name:        blueprints.BuildingName(reg.Name),
			Completed:   make(map[uuid.UUID]Building),
			Queue:       make(map[uuid.UUID]Building),
		}

		for _, building := range reg.Completed {
			b, err := restoreBuilding(building, blueprintID)
			if err != nil {
				return nil, err
			}

			register.Completed[b.ID] = b
		}

		for _, building := range reg.Queued {
			b, err := restoreBuilding(building, blueprintID)
			if err != nil {
				return nil, err
			}

			register.Queue[b.ID] = b
		}

		buildings[blueprintID] = register
	}

	return buildings, nil
}

func restoreReservedResources(pb []*protobuf.ReservedResource) []ReservedResource {
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

func restoreBuildingTimers(pb []*protobuf.BuildingTimer) ([]uuid.UUID, error) {
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
	b := Building{}

	buildingID, err := uuid.Parse(building.ID)
	if err != nil {
		return b, err
	}

	transformers, err := restoreBuildingTimers(building.Transformers)
	if err != nil {
		return b, err
	}

	generators, err := restoreBuildingTimers(building.Generators)
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
		ReservedResources: restoreReservedResources(building.ReservedResources),
		Timers: &TimerRegister{
			mx: &sync.Mutex{},

			Transformers: transformers,
			Generators:   generators,
		},
	}

	return b, nil
}

func restoreResourceRegisters(pb []*protobuf.InventoryResourceRegistry) (map[blueprints.ResourceName]*ResourceRegister, error) {
	register := make(map[blueprints.ResourceName]*ResourceRegister)

	for _, r := range pb {
		resource := &ResourceRegister{
			mx: &sync.Mutex{},

			Name:       blueprints.ResourceName(r.Name),
			CapFormula: r.CapFormula,
			Cap:        int(r.Cap),
			Amount:     int(r.Amount),
			Reserved:   int(r.Reserved),
		}

		register[resource.Name] = resource
	}

	return register, nil
}

func restoreTimers(pb []*protobuf.InventoryTimer) (map[uuid.UUID]struct{}, error) {
	timers := make(map[uuid.UUID]struct{})

	for _, t := range pb {
		id, err := uuid.Parse(t.ID)
		if err != nil {
			return nil, err
		}

		timers[id] = struct{}{}
	}

	return timers, nil
}
