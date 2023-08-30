package main

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/game"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry/remote"
	"github.com/alecthomas/kong"
	"github.com/spf13/viper"
)

type Context struct {
	Debug bool
}

type LoadCmd struct {
	Path string `arg:"" name:"path" help:"Path to the blueprint files" type:"path"`
}

func (l *LoadCmd) Run(ctx *Context) error {
	fmt.Println("loading items from " + l.Path)

	buildings, err := registry.ReadYaml[*blueprints.Building](l.Path)
	if err != nil {
		return err
	}

	for _, building := range buildings {
		building.ID = game.GetBuildingID(building.Name.String())
		if err := remote.Push(building); err != nil {
			return err
		}
	}

	resources, err := registry.ReadYaml[*blueprints.Resource](l.Path)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		if err := remote.Push(resource); err != nil {
			return err
		}
	}

	return nil
}

type ListCmd struct {
	Format string `name:"format" help:"Output format" enum:"json,yaml" default:"json"`
}

func (l *ListCmd) Run(ctx *Context) error {
	fmt.Println(viper.GetString(config.PG_Host))
	fmt.Println("Listing items in " + l.Format + " format")

	return nil
}

var CLI struct {
	Debug      bool   `help:"Enable debug mode."`
	ConfigPath string `name:"config-file" help:"Path to the config file" type:"path" default:"/etc/avalond/config.yaml"`

	Load LoadCmd `cmd:"" help:"Load blueprint files into storage"`
	List ListCmd `cmd:"" help:"List blueprints"`
}

func main() {
	ctx := kong.Parse(&CLI)

	config.Setup(CLI.ConfigPath)

	err := ctx.Run(&Context{Debug: CLI.Debug})
	ctx.FatalIfErrorf(err)
}
