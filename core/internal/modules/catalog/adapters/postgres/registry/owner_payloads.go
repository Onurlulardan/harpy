package registry

import (
	"bytes"
	"encoding/json"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func encodeOwnerPayloads(
	payloads []catalogpackagev1.OwnerPayload,
) ([]byte, error) {
	document, err := json.Marshal(payloads)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.Canonicalize(
		document, catalogpackagev1.MaximumPackageSetBytes,
	)
}

func scanPublishedVersion(source row) (packageregistry.SealedVersion, error) {
	var value packageregistry.SealedVersion
	var payloadDocument []byte
	err := source.Scan(
		&value.PackageID, &value.Slug, &value.Version, &value.Manifest,
		&value.ManifestSHA256, &value.Signature, &value.KeyID, &value.ReleaseID,
		&value.RegisteredAt, &payloadDocument,
	)
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
	canonical, err := canonicaljsonv1.Canonicalize(
		payloadDocument, catalogpackagev1.MaximumPackageSetBytes,
	)
	var stored []storedOwnerPayload
	if err != nil || json.Unmarshal(canonical, &stored) != nil {
		return packageregistry.SealedVersion{}, packageregistry.ErrUnavailable
	}
	value.OwnerPayloads = make([]catalogpackagev1.OwnerPayload, len(stored))
	for index := range stored {
		if stored[index].ReleaseID != value.ReleaseID ||
			!stored[index].RegisteredAt.UTC().Equal(value.RegisteredAt) {
			return packageregistry.SealedVersion{}, packageregistry.ErrUnavailable
		}
		value.OwnerPayloads[index] = stored[index].OwnerPayload
		value.OwnerPayloads[index].Document, err =
			canonicaljsonv1.CanonicalizeObject(
				value.OwnerPayloads[index].Document,
				catalogpackagev1.MaximumPayloadBytes,
			)
		if err != nil {
			return packageregistry.SealedVersion{}, err
		}
	}
	return value, nil
}

type storedOwnerPayload struct {
	catalogpackagev1.OwnerPayload
	ReleaseID    string    `json:"release_id"`
	RegisteredAt time.Time `json:"registered_at"`
}

func sameOwnerPayloads(
	left, right []catalogpackagev1.OwnerPayload,
) bool {
	if len(left) != len(right) {
		return false
	}
	rightByComponent := make(map[string]catalogpackagev1.OwnerPayload, len(right))
	for _, payload := range right {
		if _, duplicate := rightByComponent[payload.ComponentID]; duplicate {
			return false
		}
		rightByComponent[payload.ComponentID] = payload
	}
	for _, payload := range left {
		other, found := rightByComponent[payload.ComponentID]
		if !found || payload.Owner != other.Owner ||
			payload.Contract != other.Contract ||
			payload.ContentSHA256 != other.ContentSHA256 ||
			!bytes.Equal(payload.Document, other.Document) {
			return false
		}
		delete(rightByComponent, payload.ComponentID)
	}
	return len(rightByComponent) == 0
}
