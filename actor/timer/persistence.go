package timer

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/0xa1-red/empires-of-avalon/persistence/encoding"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/asynkron/protoactor-go/cluster"
	"google.golang.org/protobuf/types/known/structpb"
)

func (g *Grain) Encode() ([]byte, error) {
	if g.timer.Amount == 0 {
		return nil, nil
	}

	encode := make(map[string]interface{})
	data := make(map[string]interface{})

	data["timer"] = g.timer
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

	if m["timer"].(*Timer).Amount > 0 {
		g.timer = m["timer"].(*Timer)
	}

	return nil
}

func (g *Grain) Kind() string {
	return "timer"
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

	g.startBuildingTimer(context.TODO())
	g.startGenerateTimer(context.TODO())
	g.startTransformTimer(context.TODO())

	return &protobuf.RestoreResponse{
		Status: protobuf.Status_OK,
		Error:  "",
	}, nil
}

func init() {
	gob.Register(&Timer{})                      // nolint
	gob.Register(&structpb.Value_StringValue{}) // nolint
}
