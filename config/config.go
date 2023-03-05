package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

var configs = []struct {
	key string
	env string
	def interface{}
}{
	// Postgres
	{PG_User, "POSTGRES_USER", "postgres"},
	{PG_Passwd, "POSTGRESS_PASSWD", ""},
	{PG_Host, "POSTGRESS_HOST", "127.0.0.1"},
	{PG_Port, "POSTGRESS_PORT", "5432"},
	{PG_DB, "POSTGRESS_DATABASE", "defaultdb"},
	{PG_SSLMode, "POSTGRESS_SSLMODE", "disable"},
	// Cluster
	{Cluster_Name, "CLUSTER_NAME", "avalond"},
	{Node_Host, "CLUSTER_NODE_HOST", "0.0.0.0"},
	{Node_Port, "CLUSTER_NODE_PORT", 0},
	// HTTP
	{HTTP_Address, "HTTP_ADDRESS", "0.0.0.0"},
	{HTTP_Port, "HTTP_PORT", "8080"},
	// Etcd
	{ETCD_Endpoints, "ETCD_ENDPOINTS", "127.0.0.1:2379"},
	{ETCD_Root, "ETCD_ROOT", "/avalond"},
	// NATS
	{NATS_Host, "NATS_HOST", "127.0.0.1"},
	{NATS_Port, "NATS_PORT", "4222"},
	{NATS_User, "NATS_USER", ""},
	{NATS_Password, "NATS_PASSWORD", ""},
}

func Setup(path string) {
	for _, values := range configs {
		viper.SetDefault(values.key, values.def)

		viper.BindEnv(values.key, fmt.Sprintf("%s_%s", envPrefix, values.env)) // nolint
	}

	if path == "" {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("/etc/avalond")
		viper.AddConfigPath("./testdata")
		viper.AddConfigPath(".")
	} else {
		viper.SetConfigFile(path)
	}

	err := viper.ReadInConfig()
	if err != nil {
		slog.Error("failed to read configuration", err)
		os.Exit(1)
	}
}
