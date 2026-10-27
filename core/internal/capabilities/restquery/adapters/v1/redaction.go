package v1query

import (
	"fmt"
	"io"
)

const v1Redaction = "[REDACTED]"

func writeV1Redaction(state fmt.State) {
	_, _ = io.WriteString(state, v1Redaction)
}

func rejectV1JSON() ([]byte, error) {
	return nil, fmt.Errorf("V1 REST Query adapter values cannot be marshaled directly")
}
