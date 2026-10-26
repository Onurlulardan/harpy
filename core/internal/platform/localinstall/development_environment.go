package localinstall

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/google/uuid"
)

func readDevelopmentEnvironmentID(path string) (uuid.UUID, error) {
	document, err := ReadRegular(path, MaximumArtifactBytes)
	if err != nil {
		return uuid.Nil, err
	}
	defer clear(document)
	var manifest runtimev1.BootstrapAuthorityManifest
	if err = json.Unmarshal(document, &manifest); err != nil ||
		runtimev1.ValidateBootstrapAuthorityManifest(manifest) != nil {
		return uuid.Nil, fmt.Errorf("bootstrap authority is invalid")
	}
	environmentID, err := uuid.Parse(manifest.EnvironmentID)
	if err != nil || environmentID.Version() != uuid.Version(7) {
		return uuid.Nil, fmt.Errorf(
			"local environment predates Phase 14; run task local:destroy before retrying")
	}
	return environmentID, nil
}

func reconcileDevelopmentManifest(
	path string,
	environmentID uuid.UUID,
) (installationv1.Manifest, error) {
	expected, err := installationv1.Build(installationv1.BuildInput{
		EnvironmentID: environmentID.String(), DeploymentMode: installationv1.DeploymentOnPremise,
		PrimaryTenant: installationv1.TenantMetadata{Name: "Harpy Local", Slug: "harpy"},
		FirstAdmin: installationv1.AdminMetadata{
			DisplayName: "Local Instance Owner", Email: "superadmin@harpy.io",
		},
	})
	if err != nil {
		return installationv1.Manifest{}, err
	}
	document, readErr := ReadRegular(path, installationv1.MaximumBytes)
	if readErr == nil {
		defer clear(document)
		actual, parseErr := installationv1.Parse(document)
		if parseErr != nil || !bytes.Equal(actual.CanonicalBytes(), expected.CanonicalBytes()) {
			return installationv1.Manifest{}, errors.New("local installation manifest changed")
		}
		return actual, nil
	}
	if !errors.Is(readErr, os.ErrNotExist) {
		if _, statErr := os.Lstat(path); !os.IsNotExist(statErr) {
			return installationv1.Manifest{}, readErr
		}
	}
	if err = WriteExclusive(path, expected.CanonicalBytes(), 0o400); err != nil {
		return installationv1.Manifest{}, err
	}
	return expected, nil
}
