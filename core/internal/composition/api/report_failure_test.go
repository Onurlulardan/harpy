package api

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestReportRuntimeErrorIncludesCause(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelError}))
	reportRuntimeError(logger)(errors.New("IAM authorization evidence read failed"))
	text := output.String()
	if !strings.Contains(text, `"error_class":"runtime"`) ||
		!strings.Contains(text, `"service":"core-api"`) ||
		!strings.Contains(text, `"cause":"IAM authorization evidence read failed"`) {
		t.Fatalf("runtime cause missing: %s", text)
	}
	for _, forbidden := range []string{"postgres://", "SELECT ", "password="} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("runtime cause leaked %q: %s", forbidden, text)
		}
	}
}
