package contract

type Persistable interface {
	Kind() string
	Identity() string
	Encode() ([]byte, error)
}

type Restorable interface {
	Decode([]byte) error
}

type Persister interface {
	Persist(Persistable) (int, error)
}

type Restorer interface {
	Restore(kind, identity string) error
}

type PersisterRestorer interface {
	Persister
	Restorer
}

var persister PersisterRestorer
