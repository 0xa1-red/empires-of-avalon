package main

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/config"
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
