package postgres

import (
	"github.com/0xa1-red/empires-of-avalon/pkg/service/persistence"
)

func (c *Conn) Persist(item persistence.Persistable) error {
	raw, err := item.Encode()
	if err != nil {
		return err
	}

	_, err = c.Exec("INSERT INTO snapshots (identity, kind, data) VALUES ($1, $2, $3)", item.GetID(), item.GetKind(), raw)
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) Restore(item persistence.Restorable) error { //nolint
	// 	d := make(map[string]interface{})

	// 	rows, err := c.Queryx("SELECT data FROM snapshots ORDER BY created_at DESC LIMIT 1")
	// 	if err != nil {
	// 		if err == sql.ErrNoRows {
	// 			return persis
	// 		}
	// 	}
	// 	err := row.MapScan(d)
	// 	spew.Dump(err)
	// 	if err := row.MapScan(d); err != nil {
	// 		spew.Dump(err)
	// 		return err
	// 	} else if err == sql.ErrNoRows {
	// 		return persistence.ErrNoSnapshotFound
	// 	}
	return persistence.ErrNoSnapshotFound
}
