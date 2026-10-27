package lifecycledelivery

import (
	"context"
	"errors"
	"testing"
)

func TestFinalizeRejectsNilRuntime(t *testing.T) {
	var runtime *Runtime
	_, err := runtime.Finalize(context.Background(), FinalizeCommand{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
