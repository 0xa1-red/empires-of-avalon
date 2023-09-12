package registry

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/game"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry/remote"
	"github.com/google/uuid"
	"github.com/spf13/viper"
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
		Err:     nil,
	}
}

func (dwe *DecoderWithError) Decode(v any) error {
	dwe.Err = nil
	if err := dwe.Decoder.Decode(v); err != nil {
		dwe.Err = err
	}

	return dwe.Err
}

func ReadYaml[T storeable](path string) ([]T, error) {
	filename := ""

	var collection []T

	switch any(collection).(type) {
	case []*blueprints.Building:
		filename = "buildings.yaml"
	case []*blueprints.Resource:
		filename = "resources.yaml"
	}

	fp, err := os.OpenFile(filepath.Join(path, filename), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer fp.Close() // nolint

	decoder := yaml.NewDecoder(fp)

	var decodeError error

	for {
		var bp T
		if err := decoder.Decode(&bp); err != nil {
			decodeError = err
			break
		}

		collection = append(collection, bp)
	}

	if decodeError.Error() != "EOF" {
		return nil, decodeError
	}

	return collection, nil
}

func GetBuilding(name blueprints.BuildingName) (*blueprints.Building, error) {
	store, err := getStore()
	if err != nil {
		return nil, err
	}

	return store.buildings.Get(name)
}

func GetResource(name string) (*blueprints.Resource, error) {
	store, err := getStore()
	if err != nil {
		return nil, err
	}

	return store.resources.Get(blueprints.ResourceName(name))
}

func GetBuildings() (map[blueprints.BuildingName]*blueprints.Building, error) {
	store, err := getStore()
	if err != nil {
		return nil, err
	}

	store.buildings.mx.Lock()
	defer store.buildings.mx.Unlock()

	return store.buildings.store, nil
}

func GetResources() (map[blueprints.ResourceName]*blueprints.Resource, error) {
	store, err := getStore()
	if err != nil {
		return nil, err
	}

	store.resources.mx.Lock()
	defer store.resources.mx.Unlock()

	return store.resources.store, nil
}

func getStore() (*store, error) {
	if registry == nil {
		registry = &store{
			buildings: newBuildingStore(),
			resources: newResourceStore(),
		}

		if viper.GetString(config.Registry_Remote_Kind) != "memory" {
			bps, err := remote.List()
			if err != nil {
				return nil, err
			}

			for _, building := range bps["building"] {
				b := building.(*blueprints.Building)
				registry.buildings.Put(b)
			}

			for _, resource := range bps["resource"] {
				r := resource.(*blueprints.Resource)
				registry.resources.Put(r)
			}
		}
	}

	return registry, nil
}

func Push[T storeable](blueprint T) error {
	store, err := getStore()
	if err != nil {
		return err
	}

	switch bp := any(blueprint).(type) {
	case *blueprints.Building:
		if bp.ID == uuid.Nil {
			bp.ID = game.GetBuildingID(bp.Name.String())
		}

		store.buildings.Put(bp)
	case *blueprints.Resource:
		store.resources.Put(bp)
	default:
		return fmt.Errorf("invalid type %T", bp)
	}

	return nil
}

func Init() error {
	_, err := getStore()
	return err
}
