package database

import (
	"fmt"
	"os"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

type Conn struct {
	*sqlx.DB
}

var connection *Conn

func CreateConnection() error {
	dsn := buildDSN()
	slog.Debug("connecting to postgres", slog.String("dsn", dsn))
	c, err := sqlx.Connect("postgres", dsn)

	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	if err := c.Ping(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	connection = &Conn{c}

	return nil
}

func Connection() *Conn {
	if connection == nil {
		if err := CreateConnection(); err != nil {
			slog.Error("failed to connect to database", err)
			os.Exit(1)
		}
	}

	return connection
}

func buildDSN() string {
	return fmt.Sprintf("host=%s dbname=%s port=%s sslmode=%s user=%s password=%s",
		viper.GetString(config.PG_Host),
		viper.GetString(config.PG_DB),
		viper.GetString(config.PG_Port),
		viper.GetString(config.PG_SSLMode),
		viper.GetString(config.PG_User),
		viper.GetString(config.PG_Passwd),
	)
}
