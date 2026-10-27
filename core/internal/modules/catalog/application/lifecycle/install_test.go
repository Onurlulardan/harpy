package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestInstallRejectsIncompleteCommand(t *testing.T) {
	_, err := (&Application{}).Install(context.Background(), platformtx.TenantScope{}, InstallRequest{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("install error = %v", err)
	}
}
