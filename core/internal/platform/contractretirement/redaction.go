package contractretirement

import (
	"errors"
	"fmt"
)

const redaction = "[contract-retirement-redacted]"

var ErrInvalid = errors.New("contract retirement value is invalid")

func writeRedaction(state fmt.State) {
	_, _ = state.Write([]byte(redaction))
}

func rejectJSON() ([]byte, error) {
	return nil, errors.New("contract retirement value cannot be marshaled directly")
}
