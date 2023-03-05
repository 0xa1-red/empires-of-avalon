package database

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

type Conn struct {
	*sqlx.DB
}

var connection *Conn

func CreateConnection() error {
	c, err := sqlx.Open("postgres", buildDSN())
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
		CreateConnection()
	}
	return connection
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
