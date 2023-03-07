package inventory

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/common"
	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/persistence/encoding"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/spf13/viper"
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

	g.buildings = m["buildings"].(map[common.BuildingName]*BuildingRegister)
	g.resources = m["resources"].(map[common.ResourceName]*ResourceRegister)

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
	}, nil
}

func init() {
	buildingRegisters := make(map[common.BuildingName]*BuildingRegister)
	resourceRegisters := make(map[common.ResourceName]*ResourceRegister)
	gob.Register(buildingRegisters)
	gob.Register(resourceRegisters)
}
