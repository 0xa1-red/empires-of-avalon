package graph

import (
	"github.com/0xa1-red/empires-of-avalon/graph/model"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/davecgh/go-spew/spew"
)

func Buildings(d *protobuf.DescribeInventoryResponse) []*model.InventoryBuilding {
	rawBuildings, ok := d.Inventory.Fields["buildings"]
	if !ok {
		return nil
	}

	buildings := make([]*model.InventoryBuilding, 0)

	for name, raw := range rawBuildings.GetStructValue().Fields {
		n := name
		amount := int(raw.GetStructValue().Fields["amount"].GetNumberValue())
		queue := int(raw.GetStructValue().Fields["queue"].GetNumberValue())
		finish := raw.GetStructValue().Fields["finish"].GetStringValue()

		building := model.InventoryBuilding{
			Name:   &n,
			Amount: &amount,
			Queue:  &queue,
			Finish: &finish,
		}

		buildings = append(buildings, &building)
	}

	return buildings
}

func Resources(d *protobuf.DescribeInventoryResponse) []*model.InventoryResource {
	rawResources, ok := d.Inventory.Fields["resources"]
	if !ok {
		return nil
	}

	resources := make([]*model.InventoryResource, 0)

	for name, raw := range rawResources.GetStructValue().Fields {
		spew.Dump(raw)
		n := name
		amount := int(raw.GetStructValue().Fields["amount"].GetNumberValue())
		cap := int(raw.GetStructValue().Fields["cap"].GetNumberValue())
		reserved := int(raw.GetStructValue().Fields["reserved"].GetNumberValue())

		building := model.InventoryResource{
			Name:     &n,
			Amount:   &amount,
			Cap:      &cap,
			Reserved: &reserved,
		}

		resources = append(resources, &building)
	}

	return resources
}
