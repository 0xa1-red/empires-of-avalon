package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/0xa1-red/empires-of-avalon/actor/inventory"
	"github.com/0xa1-red/empires-of-avalon/actor/timer"
	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/database"
	"github.com/0xa1-red/empires-of-avalon/gamecluster"
	"github.com/0xa1-red/empires-of-avalon/logging"
	"github.com/0xa1-red/empires-of-avalon/persistence"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/automanaged"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"golang.org/x/exp/slog"
)

const testID = "e85d91f4-e56f-4ebc-9be8-c0eb107ceed0"

func main() {
	config.Setup()
	logging.Setup()

	if err := database.CreateConnection(); err != nil {
		slog.Error("failed to connect to database", err)
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	system := actor.NewActorSystem()
	provider := automanaged.New()
	lookup := disthash.New()
	remoteConfig := remote.Configure("localhost", 0)

	inventoryKind := protobuf.NewInventoryKind(func() protobuf.Inventory {
		return &inventory.Grain{}
	}, 0)

	timerKind := protobuf.NewTimerKind(func() protobuf.Timer {
		return &timer.Grain{}
	}, 0)

	clusterConfig := cluster.Configure("vslice-1", provider, lookup, remoteConfig,
		cluster.WithKinds(inventoryKind, timerKind))

	c := cluster.New(system, clusterConfig)
	c.StartMember()
	gamecluster.SetC(c)
	persistence.Create(c)

	if err := persistence.Get().Restore("inventory", ""); err != nil {
		slog.Error("failed to restore some inventory actors", err)
	}

	if err := persistence.Get().Restore("timer", ""); err != nil {
		slog.Error("failed to restore some timer actors", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go startServer(wg, ":8080")

MainLoop:
	for {
		select {
		case <-sigs:
			break MainLoop
		}
	}

	server.Shutdown(context.Background())
	wg.Wait()
	c.Shutdown(true)
}
