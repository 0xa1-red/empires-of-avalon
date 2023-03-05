package blueprints

import "github.com/google/uuid"

type Building struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Cost      map[string]int64
	Generates map[string]Generator
	BuildTime string
}

type Generator struct {
	Name       string
	Amount     int64
	TickLength string
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
