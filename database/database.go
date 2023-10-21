package database

import (
	"github.com/0xa1-red/empires-of-avalon/database/postgres"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/persistence"
	_ "github.com/lib/pq"
)

type Connection interface {
	persistence.PersistenceStore
}

var connection Connection

var _ Connection = (*postgres.Conn)(nil)

func Open() error {
	c, err := postgres.CreateConnection()
	if err != nil {
		return err
	}

	connection = c

	return nil
}

func Get() (Connection, error) {
	if connection == nil {
		if err := Open(); err != nil {
			return nil, err
		}
	}

	return connection, nil
}
