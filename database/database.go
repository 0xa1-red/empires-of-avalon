package database

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/database/pg"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

type Connection interface{}

var connection Connection

func Open() (Connection, error) {
	if connection == nil {
		kind := viper.GetString(config.DB_Kind)
		switch kind {
		case "postgres":
			c, err := pg.CreateConnection()
			if err != nil {
				return nil, err
			}

			connection = c
		default:
			return nil, fmt.Errorf("invalid database kind %s", kind)
		}
	}

	return connection, nil
}
