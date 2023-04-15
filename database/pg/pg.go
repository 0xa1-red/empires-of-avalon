package pg

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

type Conn struct {
	*sqlx.DB
}

func CreateConnection() (*Conn, error) {
	dsn := buildDSN()
	slog.Debug("connecting to postgres", slog.String("dsn", dsn))
	c, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := c.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Conn{c}, nil
}

func buildDSN() string {
	user := viper.GetString(config.PG_User)
	if pwd := viper.GetString(config.PG_Passwd); pwd != "" {
		user = fmt.Sprintf("%s:%s", user, pwd)
	}
	return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s",
		user,
		viper.GetString(config.PG_Host),
		viper.GetString(config.PG_Port),
		viper.GetString(config.PG_DB),
		viper.GetString(config.PG_SSLMode),
	)
}
