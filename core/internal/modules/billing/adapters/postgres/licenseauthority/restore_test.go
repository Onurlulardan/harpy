package licenseauthority

import (
	"context"
	"errors"
	"testing"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

func TestCommitRestoredLicenseRejectsUnboundSnapshot(t *testing.T) {
	storage, err := NewStore(DatabaseAccess{QueryRow: func(
		context.Context, string, ...any,
	) Row {
		t.Fatal("database called")
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = storage.CommitRestoredLicense(
		t.Context(), licenseapp.RestoredLicenseCommit{},
	)
	if !errors.Is(err, licenseapp.ErrAuthorityState) {
		t.Fatalf("error = %v", err)
	}
}
