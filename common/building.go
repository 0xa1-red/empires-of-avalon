package common

import "github.com/0xa1-red/empires-of-avalon/blueprints"

type BuildingName string

const (
	House      BuildingName = "house"
	Warehouse  BuildingName = "warehouse"
	Woodcutter BuildingName = "woodcutter"
	Lumberyard BuildingName = "lumberyard"
)

var Buildings map[BuildingName]Building = map[BuildingName]Building{
	House: {
		Name:      House,
		BuildTime: "10s",
		Cost: []*ResourceCost{
			{Resource: Wood, Amount: 20, Permanent: true},
			{Resource: Population, Amount: 2, Permanent: false},
		},
		Generators: []blueprints.Generator{
			{
				Name:       string(Population),
				Amount:     1,
				TickLength: "2s",
			},
		},
		Limits: map[ResourceName]int{
			Population: 6,
		},
	},

	Warehouse: {
		Name:      Warehouse,
		BuildTime: "10s",
		Cost: []*ResourceCost{
			{Resource: Wood, Amount: 50, Permanent: true},
			{Resource: Population, Amount: 5, Permanent: false},
		},
		Limits: map[ResourceName]int{
			Wood:   100,
			Stone:  100,
			Planks: 40,
		},
	},

	Woodcutter: {
		Name:      Woodcutter,
		BuildTime: "10s",
		Cost: []*ResourceCost{
			{Resource: Wood, Amount: 30, Permanent: true},
			{Resource: Population, Amount: 3, Permanent: false},
		},
		Generators: []blueprints.Generator{
			{
				Name:       string(Wood),
				Amount:     3,
				TickLength: "20s",
			},
		},
	},

	Lumberyard: {
		Name:      Lumberyard,
		BuildTime: "1s",
		Cost: []*ResourceCost{
			{Resource: Wood, Amount: 1, Permanent: false},
		},
		Transformers: []blueprints.Transformer{
			{
				Name: string(Planks),
				Cost: []blueprints.TransformerCost{
					{
						Resource:  string(wood.Name),
						Amount:    5,
						Temporary: false,
					},
				},
				Result: []blueprints.TransformerResult{
					{
						Resource: string(planks.Name),
						Amount:   1,
					},
				},
				TickLength: "10s",
			},
		},
	},
}

type Building struct {
	Name         BuildingName
	BuildTime    string
	Cost         []*ResourceCost
	Limits       map[ResourceName]int
	Generators   []blueprints.Generator
	Transformers []blueprints.Transformer
}
