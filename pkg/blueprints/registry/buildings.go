package registry

import (
	"fmt"
	"sync"

	"github.com/0xa1-red/empires-of-avalon/pkg/blueprints"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

type BuildingStore struct {
	mx *sync.Mutex

	store map[uuid.UUID]*blueprints.Building
}

func newBuildingStore() *BuildingStore {
	return &BuildingStore{
		mx:    &sync.Mutex{},
		store: make(map[uuid.UUID]*blueprints.Building),
	}
}

func (s *BuildingStore) Put(item *blueprints.Building) {
	s.mx.Lock()
	defer s.mx.Unlock()

	slog.Debug("registering building blueprint", "id", item.ID.String(), "name", item.Name.String())

	s.store[item.ID] = item
}

func (s *BuildingStore) Get(id uuid.UUID) (*blueprints.Building, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	item, ok := s.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	return item, nil
}
