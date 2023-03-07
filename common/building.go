package common

type BuildingName string

const (
	House BuildingName = "house"
)

var Buildings map[BuildingName]Building = map[BuildingName]Building{
	House: {Name: House, BuildTime: "10s", Cost: []*ResourceCost{
		{Resource: Wood, Amount: 20, Permanent: true},
		{Resource: Population, Amount: 5, Permanent: false},
	}},
}

type Building struct {
	Name      BuildingName
	BuildTime string
	Cost      []*ResourceCost
}
