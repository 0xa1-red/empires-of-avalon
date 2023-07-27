package blueprints

import (
	"bytes"
	"encoding/json"
)

type ResourceName string

func (r ResourceName) String() string {
	return string(r)
}

const (
	Population ResourceName = "Population"
	Wood       ResourceName = "Wood"
	Stone      ResourceName = "Stone"
	Planks     ResourceName = "Planks"
)

type Resource struct {
	Name           ResourceName `json:"name" yaml:"name"`
	StartingAmount int          `json:"starting_amount" yaml:"starting_amount"`
	StartingCap    int          `json:"starting_cap" yaml:"starting_cap"`
	CapFormula     string       `json:"cap_formula" yaml:"cap_formula"`
}

func (r *Resource) Encode() ([]byte, error) {
	buf := bytes.NewBuffer([]byte(""))
	encoder := json.NewEncoder(buf)

	if err := encoder.Encode(*r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *Resource) Decode(src []byte) error {
	return nil
}

func (r *Resource) Kind() string {
	return KindResource
}

func (r *Resource) GetID() string {
	return r.Name.String()
}

type ResourceCost struct {
	Resource  ResourceName
	Amount    int
	Permanent bool
}
