package nats

import (
	"fmt"
	"strings"
	"time"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/encoders/protobuf"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

var nc *nats.EncodedConn

func GetConnection() (*nats.EncodedConn, error) {
	if nc == nil {
		url := buildURL()

		slog.Debug("connecting to NATS transport layer", "url", url.String())

		c, err := nats.Connect(url.connect(), nats.Timeout(10*time.Second))
		if err != nil {
			return nil, err
		}

		enc, err := nats.NewEncodedConn(c, protobuf.PROTOBUF_ENCODER)
		if err != nil {
			return nil, err
		}

		nc = enc
	}

	return nc, nil
}

type natsURL struct {
	user     string
	password string
	host     string
	port     string
}

func buildURL() natsURL {
	return natsURL{
		user:     viper.GetString(config.NATS_User),
		password: viper.GetString(config.NATS_Password),
		host:     viper.GetString(config.NATS_Host),
		port:     viper.GetString(config.NATS_Port),
	}
}

func (n natsURL) connect() string {
	url := fmt.Sprintf("%s:%s", n.host, n.port)

	if auth := n.user; auth != "" {
		if passwd := n.password; passwd != "" {
			auth = fmt.Sprintf("%s:%s", auth, passwd)
		}

		url = fmt.Sprintf("%s@%s", auth, url)
	}

	return fmt.Sprintf("nats://%s", url)
}

func (n natsURL) String() string {
	url := n.connect()
	if n.password != "" {
		url = strings.ReplaceAll(url, n.password, "*****")
	}

	return url
}
