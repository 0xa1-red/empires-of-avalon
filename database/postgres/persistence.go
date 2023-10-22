package postgres

import (
	"database/sql"

	"github.com/0xa1-red/empires-of-avalon/database/model"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/persistence"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

func (c *Conn) Persist(item persistence.Persistable) error {
	raw, err := item.Encode()
	if err != nil {
		return err
	}

	record := model.PersistenceRecord{
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

	d := model.PersistenceRecord{}

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

func (c *Conn) GetRestorables(kind string) (map[uuid.UUID]model.PersistenceRecord, error) {
	d := []model.PersistenceRecord{}

	restorables := make(map[uuid.UUID]model.PersistenceRecord)

	if err := c.Select(&d, `SELECT DISTINCT ON (kind, identity) kind, identity, data, created_at FROM snapshots
WHERE kind = $1
ORDER BY kind, identity, created_at DESC`, kind); err != nil {
		if err == sql.ErrNoRows {
			return nil, persistence.ErrNoSnapshotFound
		}

		return nil, err
	}

	for _, record := range d {
		id, err := uuid.Parse(record.Identity)
		if err != nil {
			slog.Warn("failed to parse actor identity", "error", err, "identity", record.Identity)
			continue
		}

		restorables[id] = record
	}

	return restorables, nil
}
