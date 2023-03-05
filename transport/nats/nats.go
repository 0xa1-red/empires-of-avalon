package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/encoders/protobuf"
)

var nc *nats.EncodedConn

func GetConnection() *nats.EncodedConn {
	if nc == nil {
		c, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			panic(err)
		}

		enc, err := nats.NewEncodedConn(c, protobuf.PROTOBUF_ENCODER)
		if err != nil {
			panic(err)
		}

		nc = enc
	}

	return nc
}
