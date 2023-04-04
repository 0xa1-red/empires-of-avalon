package main

import (
	"flag"

	"github.com/0xa1-red/empires-of-avalon/blueprints"
	"github.com/0xa1-red/empires-of-avalon/blueprints/registry"
	"github.com/0xa1-red/empires-of-avalon/common"
	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/logging"
)

var (
	configPath string
)

func main() {
	flag.StringVar(&configPath, "config-file", "", "path to config file")
	flag.Parse()

	config.Setup(configPath)
	logging.Setup()

	buildingName := "house"
	id := common.GetBuildingID(buildingName)
	i := &blueprints.Building{
		ID:   id,
		Name: buildingName,
		Cost: map[string]int64{
			"wood": 100,
		},
		Generates: map[string]blueprints.Generator{
			"pops": {
				Name:       "pops",
				Amount:     1,
				TickLength: "1s",
			},
		},
		BuildTime: "10s",
	}

	if err := registry.Push("building", i, true); err != nil {
		panic(err)
	}

}
