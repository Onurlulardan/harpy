package lifecycledelivery

import (
	"context"
	"errors"
	"testing"
)

func TestReleaseRejectsNilRuntime(t *testing.T) {
	var runtime *Runtime
	err := runtime.Release(context.Background(), EffectReference{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
