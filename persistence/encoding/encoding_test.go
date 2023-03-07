package encoding_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/persistence/encoding"
	"github.com/spf13/viper"
)

func TestEncoding(t *testing.T) {
	tests := []struct {
		data interface{}
		kind string
	}{
		{
			data: map[string]interface{}{"foo": "bar"},
			kind: encoding.EncoderGob,
		},
		{
			data: map[string]interface{}{"foo": "bar"},
			kind: encoding.EncoderJSON,
		},
	}

	for i, tt := range tests {
		tf := func(t *testing.T) {
			viper.Set(config.Persistence_Encoding, tt.kind)
			buf := bytes.NewBuffer([]byte(""))

			if err := encoding.Encode(tt.data, buf); err != nil {
				t.Fatalf("FAIL: expected no errors while encoding, got %v", err)
			}

			decoded := make(map[string]interface{})
			if err := encoding.Decode(buf.Bytes(), decoded); err != nil {
				t.Fatalf("FAIL: expected no errors while decoding, got %v", err)
			}
		}
		t.Run(fmt.Sprintf("Case_%d_%s", i, tt.kind), tf)
	}
}
