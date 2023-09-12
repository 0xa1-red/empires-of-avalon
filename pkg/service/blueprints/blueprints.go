package blueprints

const (
	KindBuilding string = "building"
	KindResource string = "resource"
)

type Blueprint interface {
	Encode() ([]byte, error)
	Decode(src []byte) error
	Kind() string
	GetID() string
	GetVersion() int
	GetName() string
}

var _ Blueprint = (*Building)(nil)
var _ Blueprint = (*Resource)(nil)
