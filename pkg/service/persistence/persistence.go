package persistence

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/database/model"
	"github.com/google/uuid"
)

type PersistenceStore interface {
	Persist(item Persistable) error
	Restore(item Restorable) error
	GetRestorables(kind string) (map[uuid.UUID]model.PersistenceRecord, error)
}

type Persistable interface {
	GetID() string
	GetKind() string
	Encode() ([]byte, error)
}

type Restorable interface {
	GetID() string
	GetKind() string
	Restore(raw []byte) error
}

var ErrNoSnapshotFound = fmt.Errorf("snapshot not found")
