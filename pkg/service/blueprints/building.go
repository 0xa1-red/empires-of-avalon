package blueprints

import (
	"bytes"
	"encoding/json"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

type BuildingName string

const (
	House      BuildingName = "house"
	Warehouse  BuildingName = "warehouse"
	Woodcutter BuildingName = "woodcutter"
	Lumberyard BuildingName = "lumberyard"
)

func (r BuildingName) String() string {
	return string(r)
}

type Building struct {
	ID             uuid.UUID            `json:"id" yaml:"id"`
	Name           BuildingName         `json:"name" yaml:"name"`
	InitialAmount  int                  `json:"initial_amount" yaml:"initial_amount"`
	WorkersMaximum int                  `json:"workers_maximum" yaml:"workers_maximum"`
	Cost           []ResourceCost       `json:"cost" yaml:"cost"`
	Generates      []Generator          `json:"generates" yaml:"generates"`
	Transforms     []Transformer        `json:"transforms" yaml:"transforms"`
	Stores         map[ResourceName]int `json:"stores" yaml:"stores"`
	BuildTime      string               `json:"build_time" yaml:"build_time"`
}

type Generator struct {
	Name       ResourceName `json:"name" yaml:"name"`
	Amount     int          `json:"amount" yaml:"amount"`
	TickLength string       `json:"tick_length" yaml:"tick_length"`
}

type TransformerCost struct {
	Resource  ResourceName `json:"resource" yaml:"resource"`
	Amount    int          `json:"amount" yaml:"amount"`
	Temporary bool         `json:"is_temporary" yaml:"is_temporary"`
}

type TransformerResult struct {
	Resource ResourceName `json:"resource" yaml:"resource"`
	Amount   int          `json:"amount" yaml:"amount"`
}

type Transformer struct {
	Name       string              `json:"name" yaml:"name"`
	Cost       []TransformerCost   `json:"cost" yaml:"cost"`
	Result     []TransformerResult `json:"result" yaml:"result"`
	TickLength string              `json:"tick_length" yaml:"tick_length"`
}

func (t Transformer) CostStructList() *structpb.ListValue {
	list := structpb.ListValue{
		Values: []*structpb.Value{},
	}

	for _, c := range t.Cost {
		s := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"resource":  structpb.NewStringValue(c.Resource.String()),
				"amount":    structpb.NewNumberValue(float64(c.Amount)),
				"temporary": structpb.NewBoolValue(c.Temporary),
			},
		}

		list.Values = append(list.Values, structpb.NewStructValue(s))
	}

	return &list
}

func (t Transformer) ResultStructList() *structpb.ListValue {
	list := structpb.ListValue{
		Values: []*structpb.Value{},
	}

	for _, c := range t.Result {
		s := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"resource": structpb.NewStringValue(c.Resource.String()),
				"amount":   structpb.NewNumberValue(float64(c.Amount)),
			},
		}

		list.Values = append(list.Values, structpb.NewStructValue(s))
	}

	return &list
}

func (b *Building) Encode() ([]byte, error) {
	buf := bytes.NewBuffer([]byte(""))
	encoder := json.NewEncoder(buf)

	if err := encoder.Encode(*b); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (b *Building) Decode(src []byte) error {
	return nil
}

func (b *Building) Kind() string {
	return KindBuilding
}

func (b *Building) GetID() string {
	return b.ID.String()
}
