package recoveryanchor

import (
	"context"
	"errors"
	"testing"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
)

func TestRuntimeCreateOptionsRejectsMissingRuntime(t *testing.T) {
	var runtime *Runtime
	_, err := runtime.Options(context.Background(), anchorapi.OptionsCommand{})
	if !errors.Is(err, anchorapi.ErrInvalid) {
		t.Fatalf("CreateOptions() error = %v, want ErrInvalid", err)
	}
}
