package common

type ResourceName string

const (
	Population ResourceName = "population"
	Wood       ResourceName = "wood"
	Stone      ResourceName = "stone"
	Planks     ResourceName = "planks"
)

var Resources map[ResourceName]Resource = map[ResourceName]Resource{
	Population: population,
	Wood:       wood,
	Planks:     {Name: Planks, StartingCap: 0},
	Stone:      {Name: Stone, StartingCap: 20},
}

type Resource struct {
	Name           ResourceName
	StartingAmount int
	StartingCap    int
	CapFormula     string
}

type ResourceCost struct {
	Resource  ResourceName
	Amount    int
	Permanent bool
}

var (
	population = Resource{
		Name:           Population,
		StartingCap:    6,
		StartingAmount: 6,
		CapFormula: `
return 6+buildings.house*6
`,
	}

	wood = Resource{
		Name:           Wood,
		StartingCap:    100,
		StartingAmount: 100,
		CapFormula: `
return 100+buildings.warehouse*100
`,
	}

	planks = Resource{
		Name:           Planks,
		StartingCap:    100,
		StartingAmount: 0,
		CapFormula: `
return 100+buildings.warehouse*100
`,
	}
)
