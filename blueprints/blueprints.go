package blueprints

const (
	KindBuilding string = "building"
)

type Blueprint interface {
	Encode() ([]byte, error)
	Decode(src []byte) error
	Kind() string
	GetID() string
}
