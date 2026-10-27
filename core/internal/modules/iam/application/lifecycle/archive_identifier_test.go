package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestArchiveRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.ArchiveIdentifier(
		context.Background(), platformtx.TenantScope{}, RevisionCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
