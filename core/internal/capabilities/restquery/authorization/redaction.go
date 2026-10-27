package authorization

import (
	"fmt"
	"io"
	"log/slog"
)

const authorizationRedaction = "[REDACTED]"

func writeAuthorizationRedaction(state fmt.State) {
	_, _ = io.WriteString(state, authorizationRedaction)
}

func authorizationLogValue() slog.Value {
	return slog.StringValue(authorizationRedaction)
}

func rejectAuthorizationJSON() ([]byte, error) {
	return nil, fmt.Errorf("REST Query authorization values cannot be marshaled directly")
}
