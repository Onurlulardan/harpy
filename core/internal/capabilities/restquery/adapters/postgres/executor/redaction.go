package executor

import (
	"fmt"
	"io"
)

const redacted = "[REDACTED]"

func writeRedaction(state fmt.State) {
	_, _ = io.WriteString(state, redacted)
}
