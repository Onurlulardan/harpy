package cursor

import (
	"fmt"
	"io"
	"log/slog"
)

func writeRedaction(state fmt.State) {
	_, _ = io.WriteString(state, redacted)
}

func redactedLogValue() slog.Value {
	return slog.StringValue(redacted)
}

func rejectJSONMarshal() ([]byte, error) {
	return nil, fmt.Errorf("REST Query cursor values cannot be marshaled directly")
}
