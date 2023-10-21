package persistence

import "fmt"

type PersistenceStore interface {
	Persist(item Persistable) error
	Restore(item Restorable) error
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
