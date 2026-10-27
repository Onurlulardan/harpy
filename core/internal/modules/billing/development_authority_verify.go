package billing

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	licenseauthority "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
	"github.com/google/uuid"
)

func verifyDevelopmentAuthority(
	root string,
	manifest installationv1.Manifest,
	now time.Time,
) error {
	rootSeed, err := readDevelopmentAuthorityFile(
		root, developmentRootSeedFile, ed25519.SeedSize,
	)
	if err != nil || len(rootSeed) != ed25519.SeedSize {
		return fmt.Errorf("local development root seed is invalid")
	}
	defer clear(rootSeed)
	signerSeed, err := readDevelopmentAuthorityFile(
		root, developmentSignerSeedFile, ed25519.SeedSize,
	)
	if err != nil || len(signerSeed) != ed25519.SeedSize {
		return fmt.Errorf("local development signer seed is invalid")
	}
	defer clear(signerSeed)
	trust, err := readDevelopmentAuthorityFile(
		root, developmentTrustBundleFile, licenseauthority.MaximumDocumentBytes,
	)
	if err != nil {
		return err
	}
	defer clear(trust)
	snapshot, err := readDevelopmentAuthorityFile(
		root, developmentSnapshotFile, licenseauthority.MaximumDocumentBytes,
	)
	if err != nil {
		return err
	}
	defer clear(snapshot)
	bundle, err := licenseauthority.ParseTrustBundle(context.Background(), trust, now)
	if err != nil {
		return fmt.Errorf("verify local license trust: %w", err)
	}
	environmentID := uuid.UUID(manifest.EnvironmentID())
	identity := localinstall.DeriveIdentity(environmentID)
	verified, err := licenseauthority.Verify(context.Background(), snapshot,
		licenseauthority.Binding{LicenseID: identity.LicenseID,
			LicenseeOrganizationID: identity.LicenseeOrganizationID,
			EnvironmentID:          environmentID,
			DeliveryKind:           licenseauthority.DeliveryDevelopment}, now, bundle)
	if err != nil || verified.Issuer() != developmentLicenseIssuer || verified.Sequence() != 1 {
		return fmt.Errorf("verify local development license: %w", err)
	}
	return nil
}
