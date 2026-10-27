package restoredlicense

import (
	"context"
	"testing"

	grantapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/restoredlicense"
)

func TestReconcileRestoredLicenseRejectsUnboundStore(t *testing.T) {
	store := &store{}
	if _, err := store.ReconcileRestoredLicense(
		context.Background(), grantapp.Input{},
	); err == nil {
		t.Fatal("unbound restored license grant store was accepted")
	}
}
