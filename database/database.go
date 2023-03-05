package database

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Conn struct {
	*sqlx.DB
}

var connection *Conn

func CreateConnection() error {
	url := os.Getenv("GAME_POSTGRES_URL")
	if url == "" {
		url = "postgres://postgres@127.0.0.1:5432/defaultdb?sslmode=disable"
	}

	c, err := sqlx.Open("postgres", url)
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
