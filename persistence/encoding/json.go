package encoding

import (
	"bytes"
	"encoding/json"
	"io"
)

func jsonEncode(data interface{}, w io.Writer) error {
	encoder := json.NewEncoder(w)

	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}

func jsonDecode(data []byte, dst map[string]interface{}) error {
	buf := bytes.NewBuffer(data)
	decoder := json.NewDecoder(buf)

	if err := decoder.Decode(&dst); err != nil {
		return err
	}

	return nil
}
