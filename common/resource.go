package common

type ResourceName string

const (
	Population ResourceName = "population"
	Wood       ResourceName = "wood"
	Stone      ResourceName = "stone"
	Planks     ResourceName = "planks"
)

var Resources map[ResourceName]Resource = map[ResourceName]Resource{
	Population: {Name: Population, StartingCap: 6},
	Wood:       {Name: Wood, StartingCap: 100},
	Planks:     {Name: Planks, StartingCap: 0},
	Stone:      {Name: Stone, StartingCap: 20},
}

type Resource struct {
	Name        ResourceName
	StartingCap int
}

type ResourceCost struct {
	Resource  ResourceName
	Amount    int
	Permanent bool
}
