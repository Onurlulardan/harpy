package lifecycledelivery

import (
	"context"
	"errors"
	"testing"
)

func TestMarkRejectsNilRuntime(t *testing.T) {
	var runtime *Runtime
	err := runtime.MarkUnknown(context.Background(), EffectReference{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
