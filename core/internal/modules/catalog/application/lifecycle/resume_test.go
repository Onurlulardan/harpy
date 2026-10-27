package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestResumeRejectsMissingOperation(t *testing.T) {
	_, err := (&Application{}).Resume(context.Background(), platformtx.TenantScope{}, uuid.Nil)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("resume error = %v", err)
	}
}
