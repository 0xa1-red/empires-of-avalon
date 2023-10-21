package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/0xa1-red/empires-of-avalon/actor/admin"
	"github.com/0xa1-red/empires-of-avalon/actor/inventory"
	"github.com/0xa1-red/empires-of-avalon/actor/timer"
	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/database"
	"github.com/0xa1-red/empires-of-avalon/instrumentation/metrics"
	"github.com/0xa1-red/empires-of-avalon/instrumentation/traces"
	"github.com/0xa1-red/empires-of-avalon/logging"
	gamecluster "github.com/0xa1-red/empires-of-avalon/pkg/cluster"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/auth"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/0xa1-red/empires-of-avalon/transport/nats"
	"github.com/0xa1-red/empires-of-avalon/version"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/etcd"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
)

var (
	configPath   string
	printVersion bool
)

func main() {
	flag.BoolVar(&printVersion, "version", false, "print version information")
	flag.StringVar(&configPath, "config-file", "", "path to config file")
	flag.Parse()

	if printVersion {
		version.Print()
		os.Exit(0)
	}

	config.Setup(configPath)

	if err := logging.Setup(); err != nil {
		slog.Error("error configuring logging facility", err)
		os.Exit(1)
	}
	defer logging.Close()

	setupInstrumentation()

	initToken()

	initDatabase()

	initTransport()

	if err := initRegistry(); err != nil {
		slog.Error("failed to initialize registry", err)
		exit(1)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	etcdConf := clientv3.Config{ // nolint
		Endpoints:   viper.GetStringSlice(config.ETCD_Endpoints),
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
		Username:    viper.GetString(config.ETCD_User),
		Password:    viper.GetString(config.ETCD_Passwd),
	}

	slog.Debug("creating etcd provider",
		"endpoints", viper.GetStringSlice(config.ETCD_Endpoints),
		"username", viper.GetString(config.ETCD_User),
	)

	provider, err := etcd.NewWithConfig(viper.GetString(config.ETCD_Root), etcdConf)
	if err != nil {
		slog.Error("failed to create etcd provider", err)
		exit(1)
	}

	lookup := disthash.New()
	actorConfig := actor.Configure(actor.WithMetricProviders(otel.GetMeterProvider()))
	system := actor.NewActorSystemWithConfig(actorConfig)

	slog.Debug("configuring remote", slog.String("host", viper.GetString(config.Node_Host)), slog.String("port", viper.GetString(config.Node_Port)))

	remoteConfig := remote.Configure(viper.GetString(config.Node_Host), viper.GetInt(config.Node_Port))
	inventoryKind := protobuf.NewInventoryKind(func() protobuf.Inventory {
		return &inventory.Grain{}
	}, 0)
	timerKind := protobuf.NewTimerKind(func() protobuf.Timer {
		return &timer.Grain{}
	}, 0)
	adminKind := protobuf.NewAdminKind(func() protobuf.Admin {
		return &admin.Grain{}
	}, 0)
	clusterConfig := cluster.Configure(viper.GetString(config.Cluster_Name), provider, lookup, remoteConfig,
		cluster.WithKinds(inventoryKind, timerKind, adminKind))

	c := cluster.New(system, clusterConfig)
	c.StartMember()
	gamecluster.SetC(c)

	adminGrainClient := protobuf.GetAdminGrainClient(c, admin.AdminID.String())
	if _, err := adminGrainClient.Start(&protobuf.Empty{}); err != nil {
		slog.Error("failed to start admin grain client", err)
		exit(1)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	bindAddress := fmt.Sprintf("%s:%s",
		viper.GetString(config.HTTP_Address),
		viper.GetString(config.HTTP_Port))
	go startServer(wg, bindAddress)

	wg.Add(1)

	go metrics.ServeMetrics(wg)
	<-sigs

	if err := server.Shutdown(context.Background()); err != nil {
		slog.Warn("failed to stop HTTP server", "error", err)
	}

	if err := metrics.Shutdown(context.Background()); err != nil {
		slog.Warn("failed to stop metrics server", "error", err)
	}

	wg.Wait()
	c.Shutdown(true)

	if err := traces.Shutdown(context.Background()); err != nil {
		slog.Warn("failed to shut trace exporter down", "error", err)
	}
}

func exit(code int) {
	if err := traces.Shutdown(context.Background()); err != nil {
		slog.Warn("failed to shut trace exporter down", "error", err)
	}

	os.Exit(code)
}

func setupInstrumentation() {
	if err := metrics.RegisterMetricsPipeline(); err != nil {
		slog.Warn("failed to register metrics pipeline", "error", err)
	}

	if err := traces.RegisterTracesPipeline(); err != nil {
		slog.Warn("failed to register traces pipeline", "error", err)
	}
}

func initToken() {
	if _, err := auth.GetToken(); err != nil {
		slog.Error("failed to get management access token", err)
		exit(1)
	}
}

func initDatabase() {
	if err := database.Open(); err != nil {
		slog.Error("failed to connect to database", err)
		exit(1)
	}
}

func initTransport() {
	if _, err := nats.GetConnection(); err != nil {
		slog.Error("failed to connect to NATS", err)
		exit(1)
	}
}

func initRegistry() error {
	return registry.Init()
}
