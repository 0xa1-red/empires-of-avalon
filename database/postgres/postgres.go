package postgres

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

type Conn struct {
	*sqlx.DB
}

func CreateConnection() (*Conn, error) {
	dsn := buildDSN()
	slog.Debug("connecting to postgres", slog.String("dsn", dsn))
	c, err := sqlx.Connect("postgres", dsn)

	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := c.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Conn{c}, nil
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
