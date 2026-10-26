package registry

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

//go:embed sql/platform/register_release_version.sql
var registerReleaseVersionSQL string

// DeployRegistrar is the package registry write adapter for core-migrator deploy.
// Runtime binders intentionally expose only packageregistry.Reader.
type DeployRow interface{ Scan(...any) error }
type DeployQueryRow func(context.Context, string, ...any) DeployRow

type DeployRegistrar struct {
	queryRow DeployQueryRow
	notFound func(error) bool
}

func NewDeployRegistrar(
	queryRow DeployQueryRow,
	notFound func(error) bool,
) (*DeployRegistrar, error) {
	if queryRow == nil || notFound == nil {
		return nil, packageregistry.ErrUnavailable
	}
	return &DeployRegistrar{queryRow: queryRow, notFound: notFound}, nil
}

func newDeployRegistrar(
	queryRow DeployQueryRow,
	notFound func(error) bool,
) *DeployRegistrar {
	return &DeployRegistrar{queryRow: queryRow, notFound: notFound}
}

func (registrar *DeployRegistrar) Register(
	ctx context.Context,
	value packageregistry.SealedVersion,
) error {
	if !registrar.usable(ctx) {
		return packageregistry.ErrUnavailable
	}
	value = value.Clone()
	if !validRegistration(value) {
		return packageregistry.ErrInvalid
	}
	payloads, err := encodeOwnerPayloads(value.OwnerPayloads)
	if err != nil {
		return packageregistry.ErrInvalid
	}
	source := registrar.queryRow(ctx, registerReleaseVersionSQL,
		value.PackageID, value.Slug, value.Version, string(value.Manifest),
		value.ManifestSHA256, value.Signature, value.KeyID, value.ReleaseID,
		value.RegisteredAt, string(payloads))
	if source == nil {
		return packageregistry.ErrUnavailable
	}
	stored, err := scanPublishedVersion(source)
	if err != nil {
		if registrar.notFound(err) {
			return packageregistry.ErrConflict
		}
		return packageregistry.ErrUnavailable
	}
	if !validVersion(stored) {
		return packageregistry.ErrUnavailable
	}
	if !samePublishedVersion(stored, value) {
		return packageregistry.ErrConflict
	}
	return nil
}

func (registrar *DeployRegistrar) usable(ctx context.Context) bool {
	return registrar != nil && registrar.queryRow != nil && registrar.notFound != nil &&
		ctx != nil && ctx.Err() == nil
}

func validRegistration(value packageregistry.SealedVersion) bool {
	if !validVersion(value) || !slugPattern.MatchString(value.Slug) {
		return false
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		value.Manifest, packagecontract.MaximumManifestBytes,
	)
	if err != nil || !bytes.Equal(canonical, value.Manifest) {
		return false
	}
	digest := sha256.Sum256(value.Manifest)
	if hex.EncodeToString(digest[:]) != value.ManifestSHA256 {
		return false
	}
	manifest, err := packagecontract.Decode(value.Manifest)
	return err == nil && manifest.PackageID.String() == value.PackageID &&
		manifest.Slug == value.Slug && manifest.Version.String() == value.Version &&
		validPayloadBindings(manifest, value.OwnerPayloads)
}

func sameVersion(left, right packageregistry.SealedVersion) bool {
	return left.PackageID == right.PackageID && left.Slug == right.Slug &&
		left.Version == right.Version && bytes.Equal(left.Manifest, right.Manifest) &&
		left.ManifestSHA256 == right.ManifestSHA256 &&
		bytes.Equal(left.Signature, right.Signature) && left.KeyID == right.KeyID &&
		left.ReleaseID == right.ReleaseID && left.RegisteredAt.Equal(right.RegisteredAt) &&
		sameOwnerPayloads(left.OwnerPayloads, right.OwnerPayloads)
}

func samePublishedVersion(
	stored, proposed packageregistry.SealedVersion,
) bool {
	if stored.ReleaseID == proposed.ReleaseID {
		return sameVersion(stored, proposed)
	}
	return stored.PackageID == proposed.PackageID &&
		stored.Slug == proposed.Slug &&
		stored.Version == proposed.Version &&
		bytes.Equal(stored.Manifest, proposed.Manifest) &&
		stored.ManifestSHA256 == proposed.ManifestSHA256 &&
		sameOwnerPayloads(stored.OwnerPayloads, proposed.OwnerPayloads)
}

func validPayloadBindings(
	manifest packagecontract.Manifest,
	payloads []catalogpackagev1.OwnerPayload,
) bool {
	if len(payloads) == 0 || len(payloads) != len(manifest.Components) {
		return false
	}
	byComponent := make(map[string]catalogpackagev1.OwnerPayload, len(payloads))
	for _, payload := range payloads {
		canonical, err := canonicaljsonv1.CanonicalizeObject(
			payload.Document, catalogpackagev1.MaximumPayloadBytes,
		)
		digest := sha256.Sum256(payload.Document)
		if err != nil || !bytes.Equal(canonical, payload.Document) ||
			hex.EncodeToString(digest[:]) != payload.ContentSHA256 {
			return false
		}
		if _, duplicate := byComponent[payload.ComponentID]; duplicate {
			return false
		}
		byComponent[payload.ComponentID] = payload
	}
	for _, component := range manifest.Components {
		payload, found := byComponent[component.ID.String()]
		if !found || payload.Owner != component.Owner ||
			payload.Contract != component.Contract ||
			payload.ContentSHA256 != component.ContentHash {
			return false
		}
		delete(byComponent, component.ID.String())
	}
	return len(byComponent) == 0
}

var _ packageregistry.Registrar = (*DeployRegistrar)(nil)
