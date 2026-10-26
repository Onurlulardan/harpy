package licenseauthority

import (
	"context"
	"errors"
	"testing"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/google/uuid"
)

func TestResolveRestoreBindingRejectsInvalidStore(t *testing.T) {
	storage := &store{}
	_, err := storage.ResolveRestoreBinding(context.Background(), uuid.Nil)
	if !errors.Is(err, licenseapp.ErrAuthorityState) {
		t.Fatalf("invalid restore binding error=%v", err)
	}
}
