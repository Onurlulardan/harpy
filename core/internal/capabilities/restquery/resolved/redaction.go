package resolved

import (
	"fmt"
	"io"
)

const valueRedaction = "[REDACTED]"

func writeRedaction(state fmt.State) {
	_, _ = io.WriteString(state, valueRedaction)
}

func rejectDirectJSONMarshal() ([]byte, error) {
	return nil, fmt.Errorf("resolved REST Query values cannot be marshaled directly")
}
