package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/game"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry/remote"
	"github.com/alecthomas/kong"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Context struct {
	Debug bool
}

type LoadCmd struct {
	Path string `arg:"" name:"path" help:"Path to the blueprint files" type:"path"`
}

func (l *LoadCmd) Run(ctx *Context) error {
	fmt.Println("loading items from " + l.Path)

	buildings, err := readYaml[*blueprints.Building](l.Path)
	if err != nil {
		return err
	}

	for _, building := range buildings {
		building.ID = game.GetBuildingID(building.Name.String())
		if err := remote.Push(building); err != nil {
			return err
		}
	}

	resources, err := readYaml[*blueprints.Resource](l.Path)
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

func readYaml[T *blueprints.Building | *blueprints.Resource](path string) ([]T, error) {
	filename := ""

	var collection []T

	switch any(collection).(type) {
	case []*blueprints.Building:
		filename = "buildings.yaml"
	case []*blueprints.Resource:
		filename = "resources.yaml"
	}

	fp, err := os.OpenFile(filepath.Join(path, filename), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer fp.Close() // nolint

	decoder := yaml.NewDecoder(fp)

	var decodeError error

	for {
		var bp T
		if err := decoder.Decode(&bp); err != nil {
			decodeError = err
			break
		}

		collection = append(collection, bp)
	}

	if decodeError.Error() != "EOF" {
		return nil, decodeError
	}

	return collection, nil
}
