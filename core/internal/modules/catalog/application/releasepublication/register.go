package releasepublication

import (
	"bytes"
	"context"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func (application *Application) Register(
	ctx context.Context,
	value packageregistry.SealedVersion,
) error {
	if application == nil || application.registrar == nil || application.keys == nil ||
		ctx == nil || ctx.Err() != nil {
		return packageregistry.ErrUnavailable
	}
	value = value.Clone()
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		value.Manifest, packagecontract.MaximumManifestBytes,
	)
	if err != nil || !bytes.Equal(canonical, value.Manifest) ||
		!validReleaseID(value.ReleaseID) || !validRegistrationTime(value.RegisteredAt) {
		return packageregistry.ErrInvalid
	}
	key, err := application.keys.Resolve(value.KeyID)
	if err != nil {
		return packageregistry.ErrInvalid
	}
	manifest, err := packagecontract.Verify(value.Manifest, packagecontract.Attestation{
		ManifestSHA256: value.ManifestSHA256,
		Signature:      value.Signature,
		KeyID:          value.KeyID,
	}, key)
	if err != nil || manifest.PackageID.String() != value.PackageID ||
		manifest.Slug != value.Slug || manifest.Version.String() != value.Version ||
		!validOwnerPayloads(manifest, value.OwnerPayloads) {
		return packageregistry.ErrInvalid
	}
	return application.registrar.Register(ctx, value)
}

func validReleaseID(raw string) bool {
	value, err := packagecontract.ParseStableID(raw)
	return err == nil && value.String() == raw
}

func validRegistrationTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%1000 == 0
}
