package lifecycledelivery

import (
	"context"
	"errors"
	"testing"
)

func TestClaimRejectsNilRuntime(t *testing.T) {
	var runtime *Runtime
	_, err := runtime.Claim(context.Background(), ClaimCommand{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
