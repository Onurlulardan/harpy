package lifecycle

import (
	"bytes"
	"context"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) loadManifest(
	ctx context.Context,
	scope platformtx.TenantScope,
	packageID uuid.UUID,
	version string,
) (packagecontract.Manifest, error) {
	if packageID == uuid.Nil || version == "" {
		return packagecontract.Manifest{}, ErrInvalid
	}
	registry, err := application.registry.Bind(ctx, scope)
	if err != nil || registry == nil {
		return packagecontract.Manifest{}, ErrUnavailable
	}
	sealed, err := registry.ResolveVersion(ctx, packageID.String(), version)
	if err != nil {
		return packagecontract.Manifest{}, ErrNotFound
	}
	return application.verifySealed(sealed)
}

func (application *Application) verifySealed(
	sealed packageregistry.SealedVersion,
) (packagecontract.Manifest, error) {
	canonical, err := canonicaljsonv1.Canonicalize(sealed.Manifest, packagecontract.MaximumManifestBytes)
	if err != nil || !bytes.Equal(canonical, sealed.Manifest) {
		return packagecontract.Manifest{}, ErrUnavailable
	}
	key, err := application.keys.Resolve(sealed.KeyID)
	if err != nil {
		return packagecontract.Manifest{}, ErrUnavailable
	}
	manifest, err := packagecontract.Verify(sealed.Manifest, packagecontract.Attestation{
		ManifestSHA256: sealed.ManifestSHA256,
		Signature:      sealed.Signature,
		KeyID:          sealed.KeyID,
	}, key)
	if err != nil || manifest.PackageID.String() != sealed.PackageID ||
		manifest.Version.String() != sealed.Version || manifest.Slug != sealed.Slug {
		return packagecontract.Manifest{}, fmt.Errorf("verify catalog package: %w", ErrUnavailable)
	}
	return manifest, nil
}
