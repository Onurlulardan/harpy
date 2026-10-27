package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestIssueRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.IssueEmailVerification(
		context.Background(), platformtx.TenantScope{}, EmailVerificationStartCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
