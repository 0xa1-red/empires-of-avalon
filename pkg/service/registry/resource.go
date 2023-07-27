package registry

import (
	"fmt"
	"sync"

	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"golang.org/x/exp/slog"
)

type ResourceStore struct {
	mx *sync.Mutex

	store map[blueprints.ResourceName]*blueprints.Resource
}

func newResourceStore() *ResourceStore {
	return &ResourceStore{
		mx:    &sync.Mutex{},
		store: make(map[blueprints.ResourceName]*blueprints.Resource),
	}
}

func (s *ResourceStore) Put(item *blueprints.Resource) {
	s.mx.Lock()
	defer s.mx.Unlock()

	slog.Debug("registering building blueprint", "name", item.Name.String())

	s.store[item.Name] = item
}

func (s *ResourceStore) Get(name blueprints.ResourceName) (*blueprints.Resource, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	item, ok := s.store[name]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	return item, nil
}
