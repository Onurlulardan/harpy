package compiler

import (
	"errors"
	"fmt"
	"io"
)

const compilerRedaction = "[REDACTED]"

func writeRedaction(state fmt.State) {
	_, _ = io.WriteString(state, compilerRedaction)
}

func rejectJSON() ([]byte, error) {
	return nil, errors.New("REST Query PostgreSQL compiler value cannot be marshaled directly")
}
