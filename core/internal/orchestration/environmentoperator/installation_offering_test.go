package environmentoperator

import (
	"testing"
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	"github.com/google/uuid"
)

func TestInstallationOfferingIgnoresLicenseTable(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	createdAt := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	request := OfferingRequest{Manifest: manifest, OperationID: uuid.New(), CreatedAt: createdAt}
	pin, err := InstallationOffering(request)
	if err != nil {
		t.Fatal(err)
	}
	if pin.Revision != 1 || pin.SHA256 != manifest.SHA256() || !pin.EffectiveAt.Equal(createdAt) ||
		pin.Source.Present || pin.Source.Sequence != 0 || pin.Source.SHA256 != "" ||
		pin.VersionID == [16]byte{} {
		t.Fatalf("offering pin = %#v", pin)
	}
	again, err := InstallationOffering(request)
	if err != nil || again.VersionID != pin.VersionID {
		t.Fatalf("offering pin drifted: %#v, %v", again, err)
	}
	request.CreatedAt = createdAt.Add(time.Nanosecond)
	if _, err := InstallationOffering(request); err == nil {
		t.Fatal("non-microsecond offering time was accepted")
	}
}
