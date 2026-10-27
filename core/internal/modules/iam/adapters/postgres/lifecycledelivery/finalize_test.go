package lifecycledelivery

import (
	"context"
	"errors"
	"testing"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
)

func TestFinalizeRejectsNilStore(t *testing.T) {
	var value *store
	_, err := value.Finalize(context.Background(), lifecycledeliveryapp.FinalizeCommand{})
	if !errors.Is(err, lifecycledeliveryapp.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
