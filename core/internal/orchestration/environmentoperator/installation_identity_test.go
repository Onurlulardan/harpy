package environmentoperator

import (
	"testing"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
)

func TestDeriveInstallationIdentityIsStableV7(t *testing.T) {
	manifest := operatorManifest(t, installationv1.DeploymentOnPremise)
	first := DeriveInstallationIdentity(manifest)
	second := DeriveInstallationIdentity(manifest)
	if !first.Valid() || first != second {
		t.Fatalf("installation identity is not stable: first=%+v second=%+v", first, second)
	}
}
