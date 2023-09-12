package main

import (
	"flag"
	"log"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/logging"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry/remote"
)

var (
	configPath string
)

func main() {
	flag.StringVar(&configPath, "config-file", "", "path to config file")
	flag.Parse()

	config.Setup(configPath)
	logging.Setup() // nolint

	if err := registry.ReadYaml[*blueprints.Resource]("../../../blueprints/resources.yaml"); err != nil {
		panic(err)
	}

	if err := registry.ReadYaml[*blueprints.Building]("../../../blueprints/buildings.yaml"); err != nil {
		panic(err)
	}

	building, err := registry.GetBuilding(blueprints.House)
	if err != nil {
		panic(err)
	}

	resource, err := registry.GetResource("Wood")
	if err != nil {
		panic(err)
	}

	if err := remote.Push(building); err != nil {
		panic(err)
	}

	if err := remote.Push(resource); err != nil {
		panic(err)
	}

	// if err := remote.Push("b", "c"); err != nil {
	// 	panic(err)
	// }

	// if err := remote.Push("c", "d"); err != nil {
	// 	panic(err)
	// }

	// log.Println("key pushed")

	if r, err := remote.Get(blueprints.KindBuilding, blueprints.House.String()); err != nil {
		panic(err)
	} else {
		log.Printf("value: %#v", r)
	}
}
