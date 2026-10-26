package config

import (
	"fmt"
	"io"
	"log/slog"
)

const iamKeyRedacted = "[REDACTED]"

type iamKeyRedaction struct{}

func (iamKeyRedaction) String() string   { return iamKeyRedacted }
func (iamKeyRedaction) GoString() string { return iamKeyRedacted }
func (iamKeyRedaction) LogValue() slog.Value {
	return slog.StringValue(iamKeyRedacted)
}
func (iamKeyRedaction) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}
func (iamKeyRedaction) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, iamKeyRedacted)
}
