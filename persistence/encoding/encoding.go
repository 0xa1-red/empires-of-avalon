package encoding

import (
	"io"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/spf13/viper"
)

const (
	EncoderGob  = "gob"
	EncoderJSON = "json"
)

func Encode(data interface{}, w io.Writer) error {
	switch viper.GetString(config.Persistence_Encoding) {
	case config.EncodingJson:
		return jsonEncode(data, w)
	case config.EncodingGob:
		fallthrough
	default:
		return gobEncode(data, w)
	}
}

func Decode(data []byte, dst map[string]interface{}) error {
	switch viper.GetString(config.Persistence_Encoding) {
	case config.EncodingJson:
		return jsonDecode(data, dst)
	case config.EncodingGob:
		fallthrough
	default:
		return gobDecode(data, dst)
	}
}
