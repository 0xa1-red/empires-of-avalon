package common

type ResourceName string

const (
	Population ResourceName = "population"
	Wood       ResourceName = "wood"
	Planks     ResourceName = "planks"
)

var Resources map[ResourceName]Resource = map[ResourceName]Resource{
	Population: {Name: Population},
	Wood:       {Name: Wood},
	Planks:     {Name: Planks},
}

type Resource struct {
	Name ResourceName
}

type ResourceCost struct {
	Resource  ResourceName
	Amount    int
	Permanent bool
}
