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
	def string
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
}

func Setup() {
	for _, values := range configs {
		viper.SetDefault(values.key, values.def)

		viper.BindEnv(values.key, fmt.Sprintf("%s_%s", envPrefix, values.env)) // nolint
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/avalond")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		slog.Error("failed to read configuration", err)
		os.Exit(1)
	}
}
