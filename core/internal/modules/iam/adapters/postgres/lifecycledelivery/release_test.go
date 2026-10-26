package lifecycledelivery

import (
	"context"
	"errors"
	"testing"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
)

func TestReleaseRejectsNilStore(t *testing.T) {
	var value *store
	err := value.Release(context.Background(), lifecycledeliveryapp.EffectReference{})
	if !errors.Is(err, lifecycledeliveryapp.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
