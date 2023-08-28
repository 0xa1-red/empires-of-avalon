package remote

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry/remote/etcd"
	"github.com/spf13/viper"
)

type backend interface {
	Push(bp blueprints.Blueprint) error
	List() (map[string]blueprints.Blueprint, error)
	Get(kind, key string) (blueprints.Blueprint, error)
}

var connection backend

func getConnection() (backend, error) {
	if connection == nil {
		switch viper.GetString(config.Registry_Remote_Kind) {
		case "etcd":
			c, err := etcd.New()
			if err != nil {
				return nil, err
			}

			connection = c
		default:
			return nil, fmt.Errorf("invalid backend kind %s", viper.GetString(config.Registry_Remote_Kind))
		}
	}

	return connection, nil
}

func Push(bp blueprints.Blueprint) error {
	c, err := getConnection()
	if err != nil {
		return err
	}

	return c.Push(bp)
}

func List() (map[string]blueprints.Blueprint, error) {
	c, err := getConnection()
	if err != nil {
		return nil, err
	}

	return c.List()
}

func Get(kind, key string) (blueprints.Blueprint, error) {
	c, err := getConnection()
	if err != nil {
		return nil, err
	}

	return c.Get(kind, key)
}
