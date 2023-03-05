package nats

import (
	"fmt"
	"log"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/encoders/protobuf"
	"github.com/spf13/viper"
)

var nc *nats.EncodedConn

func GetConnection() *nats.EncodedConn {
	if nc == nil {
		c, err := nats.Connect(buildURL())
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

func buildURL() string {
	url := fmt.Sprintf("%s:%s", viper.GetString(config.NATS_Host), viper.GetString(config.NATS_Port))
	if auth := viper.GetString(config.NATS_User); auth != "" {
		log.Println(auth)
		if passwd := viper.GetString(config.NATS_Password); passwd != "" {
			auth = fmt.Sprintf("%s:%s", auth, passwd)
		}
		url = fmt.Sprintf("%s@%s", auth, url)
	}

	return fmt.Sprintf("nats://%s", url)
}
