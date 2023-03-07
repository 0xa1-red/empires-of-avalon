package common

import (
	"encoding/gob"
)

func init() {
	gob.Register(Buildings)
	gob.Register(Resources)
}
