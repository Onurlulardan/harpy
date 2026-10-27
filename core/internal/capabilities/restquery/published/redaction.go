package published

import (
	"fmt"
	"io"
)

const publishedRedaction = "[REDACTED]"

func writePublishedRedaction(state fmt.State) {
	_, _ = io.WriteString(state, publishedRedaction)
}

func rejectPublishedJSON() ([]byte, error) {
	return nil, fmt.Errorf("published REST Query values cannot be marshaled directly")
}
