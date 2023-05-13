package blueprints

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

type Building struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Cost       map[string]int64
	Generates  map[string]Generator
	Transforms map[string]Transformer
	BuildTime  string
}

type Generator struct {
	Name       string
	Amount     int
	TickLength string
}

type TransformerCost struct {
	Resource  string
	Amount    int
	Temporary bool
}

type TransformerResult struct {
	Resource string
	Amount   int
}

type Transformer struct {
	Name       string
	Cost       []TransformerCost
	Result     []TransformerResult
	TickLength string
}

func (t Transformer) CostStructList() *structpb.ListValue {
	list := structpb.ListValue{
		Values: []*structpb.Value{},
	}

	for _, c := range t.Cost {
		s := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"resource":  structpb.NewStringValue(c.Resource),
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
				"resource": structpb.NewStringValue(c.Resource),
				"amount":   structpb.NewNumberValue(float64(c.Amount)),
			},
		}

		list.Values = append(list.Values, structpb.NewStructValue(s))
	}

	return &list
}

func (b *Building) Encode() ([]byte, error) {
	return nil, nil
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
