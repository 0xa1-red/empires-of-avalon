package common

import (
	"encoding/gob"
)

type BuildingName string

const (
	House BuildingName = "house"
)

var Buildings map[BuildingName]Building = map[BuildingName]Building{
	House: {Name: House, BuildTime: "10s"},
}

type Building struct {
	Name      BuildingName
	BuildTime string
}

func init() {
	gob.Register(Buildings)
}
