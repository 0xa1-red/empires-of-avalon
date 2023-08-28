package inventory

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/persistence/encoding"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/registry"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

func (g *Grain) Encode() ([]byte, error) {
	encode := make(map[string]interface{})
	data := make(map[string]interface{})

	data["buildings"] = g.buildings
	data["resources"] = g.resources
	encode["data"] = data
	encode["identity"] = g.ctx.Identity()

	buf := bytes.NewBuffer([]byte(""))

	if err := encoding.Encode(data, buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *Grain) Decode(b []byte) error {
	m := make(map[string]interface{})

	if err := encoding.Decode(b, m); err != nil {
		return err
	}

	if viper.GetString(config.Persistence_Encoding) == config.EncodingJson {
		return fmt.Errorf("Unimplemented")
	}

	g.buildings = m["buildings"].(map[uuid.UUID]*BuildingRegister)
	g.resources = m["resources"].(map[blueprints.ResourceName]*ResourceRegister)

	for _, r := range g.resources {
		r.mx = &sync.Mutex{}
	}

	for blueprintID, b := range g.buildings {
		b.mx = &sync.Mutex{}
		if len(b.Completed) == 0 {
			continue
		}

		slog.Debug("building decode", "name", b.Name, "amount", len(b.Completed))

		blueprint, err := registry.GetBuilding(b.Name)
		if err != nil {
			slog.Error("failed to retrieve blueprint from registry", err, "blueprint_id", blueprintID)
		}

		for buildingID := range b.Completed {
			g.startBuildingGenerators(buildingID, blueprint)
			g.startBuildingTransformers(buildingID, blueprint)
		}
	}

	g.updateLimits()

	return nil
}

func (g *Grain) Kind() string {
	return "inventory"
}

func (g *Grain) Identity() string {
	return g.ctx.Identity()
}

func (g *Grain) Restore(req *protobuf.RestoreRequest, ctx cluster.GrainContext) (*protobuf.RestoreResponse, error) {
	if err := g.Decode(req.Data); err != nil {
		return &protobuf.RestoreResponse{
			Status: protobuf.Status_Error,
			Error:  err.Error(),
		}, nil
	}

	return &protobuf.RestoreResponse{
		Status: protobuf.Status_OK,
		Error:  "",
	}, nil
}

func init() {
	buildingRegisters := make(map[blueprints.BuildingName]*BuildingRegister)
	resourceRegisters := make(map[blueprints.ResourceName]*ResourceRegister)

	gob.Register(buildingRegisters)
	gob.Register(resourceRegisters)
}
