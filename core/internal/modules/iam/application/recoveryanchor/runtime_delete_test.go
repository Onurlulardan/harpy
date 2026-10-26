package recoveryanchor

import (
	"context"
	"errors"
	"testing"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
)

func TestRuntimeDeleteRejectsMissingRuntime(t *testing.T) {
	var runtime *Runtime
	_, err := runtime.ProposeDeletion(context.Background(), anchorapi.DeleteCommand{})
	if !errors.Is(err, anchorapi.ErrInvalid) {
		t.Fatalf("Delete() error = %v, want ErrInvalid", err)
	}
}
