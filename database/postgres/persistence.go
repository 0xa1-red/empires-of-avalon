package postgres

import (
	"database/sql"
	"time"

	"github.com/0xa1-red/empires-of-avalon/pkg/service/persistence"
)

type PersistenceRecord struct {
	Kind      string    `db:"kind"`
	Identity  string    `db:"identity"`
	Data      []byte    `db:"data"`
	CreatedAt time.Time `db:"created_at"`
}

func (c *Conn) Persist(item persistence.Persistable) error {
	raw, err := item.Encode()
	if err != nil {
		return err
	}

	record := PersistenceRecord{
		Kind:     item.GetKind(),
		Identity: item.GetID(),
		Data:     raw,
	}

	_, err = c.NamedExec("INSERT INTO snapshots (identity, kind, data) VALUES (:identity, :kind, :data)", &record)
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) Restore(item persistence.Restorable) error { //nolint

	d := PersistenceRecord{}

	if err := c.QueryRowx(`SELECT * FROM snapshots
WHERE kind = $1 AND identity = $2
ORDER BY created_at DESC
LIMIT 1`, item.GetKind(), item.GetID()).StructScan(&d); err != nil {
		if err == sql.ErrNoRows {
			return persistence.ErrNoSnapshotFound
		}

		return err
	}

	return item.Restore(d.Data)
}
