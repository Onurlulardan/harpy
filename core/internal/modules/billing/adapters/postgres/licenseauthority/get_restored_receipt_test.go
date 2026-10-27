package licenseauthority

import (
	"context"
	"errors"
	"testing"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/google/uuid"
)

func TestGetRestoredLicenseReceiptRejectsInvalidBinding(t *testing.T) {
	storage := &store{}
	_, err := storage.GetRestoredLicenseReceipt(
		context.Background(), uuid.Nil, uuid.Nil,
	)
	if !errors.Is(err, licenseapp.ErrAuthorityState) {
		t.Fatalf("invalid restored receipt read error=%v", err)
	}
}
