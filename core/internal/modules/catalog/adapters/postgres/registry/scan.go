package registry

import (
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func scanVersion(source row) (packageregistry.SealedVersion, error) {
	var value packageregistry.SealedVersion
	err := source.Scan(&value.PackageID, &value.Slug, &value.Version, &value.Manifest,
		&value.ManifestSHA256, &value.Signature, &value.KeyID, &value.ReleaseID,
		&value.RegisteredAt)
	if err != nil {
		return packageregistry.SealedVersion{}, err
	}
	value.Manifest, err = canonicaljsonv1.CanonicalizeObject(
		value.Manifest, packagecontract.MaximumManifestBytes,
	)
	if err != nil {
		return packageregistry.SealedVersion{}, err
	}
	value.RegisteredAt = value.RegisteredAt.UTC()
	return value, nil
}

func validVersion(value packageregistry.SealedVersion) bool {
	return validPackageID(value.PackageID) && slugPattern.MatchString(value.Slug) &&
		storedVersionPattern.MatchString(value.Version) &&
		len(value.Manifest) > 0 && digestPattern.MatchString(value.ManifestSHA256) &&
		len(value.Signature) == 64 && value.KeyID != "" && len(value.KeyID) <= 128 &&
		validStableID(value.ReleaseID) &&
		!value.RegisteredAt.IsZero() && value.RegisteredAt.Location() == time.UTC &&
		value.RegisteredAt.Nanosecond()%1000 == 0
}
