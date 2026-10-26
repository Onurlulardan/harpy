package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestInstallationByIDRejectsNonV7Identity(t *testing.T) {
	_, err := (&Application{}).InstallationByID(
		context.Background(), platformtx.TenantScope{}, uuid.New())
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("get installation by id error = %v", err)
	}
}
