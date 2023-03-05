package config

import (
	"os"

	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

const (
	PG_User    = "Postgres.User"
	PG_Passwd  = "Postgres.Password"
	PG_Host    = "Postgres.Host"
	PG_Port    = "Postgres.Port"
	PG_DB      = "Postgres.Database"
	PG_SSLMode = "Postgres.SSLMode"
)

func Setup() {
	//postgres://postgres@127.0.0.1:5432/defaultdb?sslmode=disable
	viper.SetDefault(PG_User, "postgres")
	viper.SetDefault(PG_Passwd, "")
	viper.SetDefault(PG_Host, "127.0.0.1")
	viper.SetDefault(PG_Port, "5432")
	viper.SetDefault(PG_DB, "defaultdb")
	viper.SetDefault(PG_SSLMode, "disable")

	viper.BindEnv(PG_User, "AVALOND_POSTGRES_USER")       // nolint
	viper.BindEnv(PG_Passwd, "AVALOND_POSTGRES_PASSWD")   // nolint
	viper.BindEnv(PG_Host, "AVALOND_POSTGRES_HOST")       // nolint
	viper.BindEnv(PG_Port, "AVALOND_POSTGRES_PORT")       // nolint
	viper.BindEnv(PG_DB, "AVALOND_POSTGRES_DATABASE")     // nolint
	viper.BindEnv(PG_SSLMode, "AVALOND_POSTGRES_SSLMODE") // nolint

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/avalond")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		slog.Error("failed to read configuration", err)
		os.Exit(1)
	}

	slog.Info("postgres config",
		slog.String("user", viper.GetString(PG_User)),
		slog.String("password", viper.GetString(PG_Passwd)),
		slog.String("host", viper.GetString(PG_Host)),
		slog.String("port", viper.GetString(PG_Port)),
		slog.String("database", viper.GetString(PG_DB)),
		slog.String("ssl-mode", viper.GetString(PG_SSLMode)),
	)
}
