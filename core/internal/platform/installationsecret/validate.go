package installationsecret

import (
	"bytes"
	"encoding/base64"
)

func validPasswordWire(wire []byte) bool {
	if len(wire) != wirePasswordBytes {
		return false
	}
	raw := make([]byte, rawPasswordBytes)
	count, err := base64.RawURLEncoding.Decode(raw, wire)
	if err != nil || count != rawPasswordBytes || bytes.Equal(raw, make([]byte, rawPasswordBytes)) {
		clear(raw)
		return false
	}
	canonical := make([]byte, wirePasswordBytes)
	base64.RawURLEncoding.Encode(canonical, raw)
	valid := bytes.Equal(canonical, wire)
	clear(canonical)
	clear(raw)
	return valid
}
