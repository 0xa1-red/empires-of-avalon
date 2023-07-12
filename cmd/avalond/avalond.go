package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

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
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/etcd"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
)

var configPath string

func main() {
	flag.StringVar(&configPath, "config-file", "", "path to config file")
	flag.Parse()

	config.Setup(configPath)
	logging.Setup()

	if err := database.CreateConnection(); err != nil {
		slog.Error("failed to connect to database", err)
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	system := actor.NewActorSystem()

	provider, err := etcd.NewWithConfig(viper.GetString(config.ETCD_Root), clientv3.Config{ // nolint
		Endpoints:   viper.GetStringSlice(config.ETCD_Endpoints),
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	})
	if err != nil {
		log.Fatalf("error creating etcd provider: %v", err)
	}

	lookup := disthash.New()

	slog.Debug("configuring remote", slog.String("host", viper.GetString(config.Node_Host)), slog.String("port", viper.GetString(config.Node_Port)))

	remoteConfig := remote.Configure(viper.GetString(config.Node_Host), viper.GetInt(config.Node_Port))
	inventoryKind := protobuf.NewInventoryKind(func() protobuf.Inventory {
		return &inventory.Grain{}
	}, 0)
	timerKind := protobuf.NewTimerKind(func() protobuf.Timer {
		return &timer.Grain{}
	}, 0)
	clusterConfig := cluster.Configure(viper.GetString(config.Cluster_Name), provider, lookup, remoteConfig,
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

	bindAddress := fmt.Sprintf("%s:%s",
		viper.GetString(config.HTTP_Address),
		viper.GetString(config.HTTP_Port))
	go startServer(wg, bindAddress)

	<-sigs

	server.Shutdown(context.Background()) // nolint
	wg.Wait()
	c.Shutdown(true)
}
