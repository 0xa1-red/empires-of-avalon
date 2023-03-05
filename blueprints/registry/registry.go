package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/0xa1-red/empires-of-avalon/blueprints"
	"github.com/0xa1-red/empires-of-avalon/database"
	"github.com/google/uuid"
)

var registry *itemRegistry

type itemCollection map[string]map[uuid.UUID]blueprints.Blueprint

type itemRegistry struct {
	mx *sync.Mutex

	cache itemCollection
	db    *database.Conn
}

func new() *itemRegistry {
	return &itemRegistry{
		mx:    &sync.Mutex{},
		cache: make(itemCollection),
		db:    database.Connection(),
	}
}

func Get(kind string, id uuid.UUID) (blueprints.Blueprint, error) {
	if registry == nil {
		registry = new()
	}

	item, err := lookupLocal(kind, id)
	if err == nil {
		return item, nil
	}

	if errors.Is(err, NotFoundError{}) {
		item, err = lookupRemote(kind, id)
		if err == nil {
			return item, nil
		}
	}

	return nil, err
}

func Push(kind string, item blueprints.Blueprint, remote ...bool) error {
	isRemote := false
	if remote != nil {
		isRemote = remote[0]
	}
	if err := push(kind, item); err != nil {
		return err
	}

	var err error
	if isRemote {
		if registry == nil {
			registry = new()
		}
		log.Printf("%v", registry.db)
		err = registry.pushBuildingBlueprintRemote(kind, item)
	}
	return err
}

func lookupLocal(kind string, id uuid.UUID) (blueprints.Blueprint, error) {
	if registry == nil {
		registry = new()
		return nil, NewNotFoundError(kind, id)
	}

	registry.mx.Lock()
	defer registry.mx.Unlock()
	if item, ok := registry.cache[kind][id]; !ok {
		return nil, NewNotFoundError(kind, id)
	} else if i, ok := item.(blueprints.Blueprint); !ok {
		return nil, NewNotFoundError(kind, id)
	} else {
		return i, nil
	}
}

func lookupRemote(kind string, id uuid.UUID) (blueprints.Blueprint, error) {
	p := database.Connection()

	r := p.QueryRowx("SELECT * FROM blueprints WHERE kind = $1 AND id = $2", kind, id.String())
	if r.Err() != nil {
		return nil, r.Err()
	}

	var res blueprints.Blueprint
	switch kind {
	case "building":
		res = &blueprints.Building{}
		if err := r.Scan(&res); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func push(kind string, item blueprints.Blueprint) error {
	if registry == nil {
		registry = new()
	}

	registry.mx.Lock()
	defer registry.mx.Unlock()

	if registry.cache == nil {
		registry.cache = make(itemCollection)
	}

	if registry.cache[kind] == nil {
		registry.cache[kind] = make(map[uuid.UUID]blueprints.Blueprint)
	}

	parsedID, err := uuid.Parse(item.GetID())
	if err != nil {
		return err
	}
	registry.cache[kind][parsedID] = item

	return nil
}

func (r *itemRegistry) pushBuildingBlueprintRemote(kind string, item blueprints.Blueprint) error {
	blueprint, ok := item.(*blueprints.Building)
	if !ok {
		return fmt.Errorf("invalid item type %T", item)
	}
	raw := bytes.NewBuffer([]byte(""))
	encoder := json.NewEncoder(raw)
	err := encoder.Encode(blueprint)
	if err != nil {
		return err
	}

	log.Printf("%v", r.db)

	if _, err := r.db.Exec("INSERT INTO blueprints (id, kind, data) VALUES ($1, $2, $3)", blueprint.ID, kind, raw.String()); err != nil {
		return err
	}

	return nil
}
