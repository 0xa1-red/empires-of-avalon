package registry

import (
	"fmt"
	"io"
	"os"

	"github.com/0xa1-red/empires-of-avalon/pkg/blueprints"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

type storeable interface {
	*blueprints.Building | *blueprints.Resource
}

type InvalidFieldError struct {
	Field string
}

func (ife InvalidFieldError) Error() string {
	return fmt.Sprintf("failed to read field: %s", ife.Field)
}

func NewInvalidFieldError(field string) InvalidFieldError {
	return InvalidFieldError{
		Field: field,
	}
}

var registry *store

type store struct {
	buildings *BuildingStore
	resources *ResourceStore
}

type DecoderWithError struct {
	*yaml.Decoder

	Err error
}

func NewDecoderWithError(r io.Reader) *DecoderWithError {
	return &DecoderWithError{
		Decoder: yaml.NewDecoder(r),
	}
}

func (dwe *DecoderWithError) Decode(v any) error {
	dwe.Err = nil
	if err := dwe.Decoder.Decode(v); err != nil {
		dwe.Err = err
	}

	return dwe.Err
}

func ReadYaml[V storeable](path string) error {
	fp, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer fp.Close()

	decoder := NewDecoderWithError(fp)

	store := getStore()

	var item V

	for decoder.Decode(&item) == nil {
		switch blueprint := any(item).(type) {
		case *blueprints.Building:
			if blueprint.ID == uuid.Nil {
				blueprint.ID = blueprints.GetBuildingID(blueprint.Name.String())
			}

			store.buildings.Put(blueprint)
		case *blueprints.Resource:
			store.resources.Put(blueprint)
		default:
			return fmt.Errorf("invalid type")
		}

		item = nil
	}

	if decoder.Err != nil && decoder.Err.Error() != "EOF" {
		slog.Error("failed to decode yaml entry", decoder.Err)
		return decoder.Err
	}

	return nil
}

func GetBuilding(id uuid.UUID) (*blueprints.Building, error) {
	store := getStore()
	return store.buildings.Get(id)
}

func GetResource(name string) (*blueprints.Resource, error) {
	store := getStore()
	return store.resources.Get(blueprints.ResourceName(name))
}

func GetBuildings() map[uuid.UUID]*blueprints.Building {
	store := getStore()
	store.buildings.mx.Lock()
	defer store.buildings.mx.Unlock()

	return store.buildings.store
}

func GetResources() map[blueprints.ResourceName]*blueprints.Resource {
	store := getStore()
	store.resources.mx.Lock()
	defer store.resources.mx.Unlock()

	return store.resources.store
}

func getStore() *store {
	if registry == nil {
		registry = &store{
			buildings: newBuildingStore(),
			resources: newResourceStore(),
		}
	}

	return registry
}
