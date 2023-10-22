package model

import "time"

type PersistenceRecord struct {
	Kind      string    `db:"kind"`
	Identity  string    `db:"identity"`
	Data      []byte    `db:"data"`
	CreatedAt time.Time `db:"created_at"`
}
