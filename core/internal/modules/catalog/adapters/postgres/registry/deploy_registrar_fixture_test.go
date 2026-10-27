package registry

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func deployVersion(t *testing.T) packageregistry.SealedVersion {
	t.Helper()
	payloadDocument, err := canonicaljsonv1.CanonicalizeObject(
		[]byte(fmt.Sprintf(
			`{"component_id":"%s","definition":{"slug":"deploy-test"}}`,
			deployComponentID,
		)),
		catalogpackagev1.MaximumPayloadBytes,
	)
	if err != nil {
		t.Fatal(err)
	}
	payloadDigest := sha256.Sum256(payloadDocument)
	payloadDigestHex := hex.EncodeToString(payloadDigest[:])
	document := []byte(fmt.Sprintf(
		`{"contract":"catalog-package.v1","package_id":"%s","slug":"deploy-test",`+
			`"version":"1.0.0","platform_contract":"harpy.v2","profile":"business",`+
			`"dependencies":[],"components":[{"ID":"%s","Owner":"forms",`+
			`"Contract":"forms.v2","kind":"FORM","content_sha256":"%s"}],"menu":[]}`,
		deployPackageID, deployComponentID, payloadDigestHex,
	))
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, packagecontract.MaximumManifestBytes,
	)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(canonical)
	return packageregistry.SealedVersion{
		PackageID: deployPackageID, Slug: "deploy-test", Version: "1.0.0",
		Manifest: canonical, ManifestSHA256: hex.EncodeToString(digest[:]),
		Signature: bytes.Repeat([]byte{0x44}, 64), KeyID: "release-2026-07",
		ReleaseID:    deployReleaseID,
		RegisteredAt: time.Date(2026, 7, 30, 12, 0, 0, 123000, time.UTC),
		OwnerPayloads: []catalogpackagev1.OwnerPayload{{
			ComponentID: deployComponentID, Owner: "forms",
			Contract: "forms.v2", ContentSHA256: payloadDigestHex,
			Document: payloadDocument,
		}},
	}
}
