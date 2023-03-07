package encoding

import (
	"bytes"
	"encoding/gob"
	"io"
)

func gobEncode(data interface{}, w io.Writer) error {
	encoder := gob.NewEncoder(w)

	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}

func gobDecode(data []byte, dst map[string]interface{}) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)

	if err := decoder.Decode(&dst); err != nil {
		return err
	}

	return nil
}
