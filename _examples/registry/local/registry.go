package main

import (
	"flag"

	"github.com/0xa1-red/empires-of-avalon/blueprints"
	"github.com/0xa1-red/empires-of-avalon/blueprints/registry"
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
	logging.Setup() // nolint

	if err := registry.ReadYaml[*blueprints.Resource]("./resources.yaml"); err != nil {
		panic(err)
	}

	if err := registry.ReadYaml[*blueprints.Building]("./buildings.yaml"); err != nil {
		panic(err)
	}
}
