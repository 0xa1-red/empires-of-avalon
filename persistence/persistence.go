package persistence

import (
	"github.com/0xa1-red/empires-of-avalon/persistence/contract"
	"github.com/0xa1-red/empires-of-avalon/persistence/postgres"
	"github.com/asynkron/protoactor-go/cluster"
)

var persister contract.PersisterRestorer

func Create(c *cluster.Cluster) {
	if persister == nil {
		persister = postgres.NewPersister(c)
	}
}

func Get() contract.PersisterRestorer {
	return persister
}
