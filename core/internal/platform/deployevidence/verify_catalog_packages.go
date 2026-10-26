package deployevidence

import (
	"fmt"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func verifyCatalogPackageSet(
	settings config.MigratorDeploy,
	authority releaseAuthority,
) (attestation.VerifiedCatalogPackageSet, error) {
	document, err := readArtifact(
		settings.Verification().CatalogPackageSetPath().Value(),
		catalogpackagev1.MaximumPackageSetBytes,
		"Catalog package set",
	)
	if err != nil {
		return attestation.VerifiedCatalogPackageSet{}, err
	}
	verified, err := attestation.VerifyCatalogPackageSet(
		document.bytes,
		authority.release,
		settings.CatalogPackageKeyring(),
		authority.root.ActivePublicKeys(authority.release.IssuedAt()),
	)
	if err != nil {
		return attestation.VerifiedCatalogPackageSet{},
			fmt.Errorf("Catalog package set verification failed")
	}
	return verified, nil
}
