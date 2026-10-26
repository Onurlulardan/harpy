package execution

import (
	"fmt"
	"io"
	"log/slog"
)

const executionRedaction = "[REDACTED]"

func writeExecutionRedaction(state fmt.State) {
	_, _ = io.WriteString(state, executionRedaction)
}

func executionLogValue() slog.Value {
	return slog.StringValue(executionRedaction)
}

func rejectExecutionJSON() ([]byte, error) {
	return nil, errExecutionContract
}
