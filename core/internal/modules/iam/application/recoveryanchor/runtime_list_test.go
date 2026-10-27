package recoveryanchor

import (
	"context"
	"errors"
	"testing"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
)

func TestRuntimeListRejectsMissingRuntime(t *testing.T) {
	var runtime *Runtime
	_, err := runtime.Anchors(context.Background(), anchorapi.Subject{})
	if !errors.Is(err, anchorapi.ErrInvalid) {
		t.Fatalf("List() error = %v, want ErrInvalid", err)
	}
}
