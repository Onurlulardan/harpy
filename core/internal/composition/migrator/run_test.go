package migrator

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func TestRunRejectsInvalidInvocationBeforeConfiguration(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(context.Background(), []string{"deploy"}, &stdout, &stderr)
	if code != processcode.InvalidStartup {
		t.Fatalf("code=%d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if !strings.Contains(stderr.String(), `"error_class":"invocation"`) ||
		!strings.Contains(stderr.String(), "usage: core-migrator") {
		t.Fatalf("stable invocation failure missing: %s", stderr.String())
	}
}
