package registry

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

const registryPackageID = "018f5000-0000-7000-8000-000000000001"
const registryComponentID = "018f5000-0000-7000-8000-000000000002"
const registryReleaseID = "018f5000-0000-7000-8000-000000000003"

type registryVersionRow struct {
	manifest, signature []byte
	digest              string
}

func (fixture registryVersionRow) Scan(destinations ...any) error {
	if len(destinations) != 9 {
		return fmt.Errorf("registry version destinations = %d", len(destinations))
	}
	*(destinations[0].(*string)) = registryPackageID
	*(destinations[1].(*string)) = "catalog-test"
	*(destinations[2].(*string)) = "1.0.0"
	*(destinations[3].(*[]byte)) = append([]byte(nil), fixture.manifest...)
	digest := fixture.digest
	if digest == "" {
		digest = "1111111111111111111111111111111111111111111111111111111111111111"
	}
	signature := fixture.signature
	if signature == nil {
		signature = bytes.Repeat([]byte{0x22}, 64)
	}
	*(destinations[4].(*string)) = digest
	*(destinations[5].(*[]byte)) = append([]byte(nil), signature...)
	*(destinations[6].(*string)) = "catalog-key-1"
	*(destinations[7].(*string)) = registryReleaseID
	*(destinations[8].(*time.Time)) = time.Date(
		2026, 7, 30, 12, 0, 0, 123000, time.FixedZone("registry-test", 3*60*60),
	)
	return nil
}

func TestScanVersionRestoresCanonicalManifestBytesFromJSONB(t *testing.T) {
	postgresJSONB := []byte(`{"version": "1.0.0", "slug": "catalog-test"}`)
	value, err := scanVersion(registryVersionRow{manifest: postgresJSONB})
	if err != nil {
		t.Fatalf("scan version: %v", err)
	}
	want := []byte(`{"slug":"catalog-test","version":"1.0.0"}`)
	if !bytes.Equal(value.Manifest, want) {
		t.Fatalf("manifest = %s want %s", value.Manifest, want)
	}
	if value.RegisteredAt.Location() != time.UTC ||
		value.RegisteredAt.Nanosecond()%1000 != 0 {
		t.Fatalf("registered_at = %s", value.RegisteredAt)
	}
	if value.ReleaseID != registryReleaseID {
		t.Fatalf("release_id = %s", value.ReleaseID)
	}
}

func TestScanVersionKeepsSignedManifestVerifiableAfterJSONBRoundTrip(t *testing.T) {
	postgresJSONB := []byte(fmt.Sprintf(
		`{"version": "1.0.0", "slug": "catalog-test", "profile": "business",`+
			` "platform_contract": "harpy.v2", "package_id": "%s", "menu": [],`+
			` "dependencies": [], "contract": "catalog-package.v1", "components": [`+
			`{"content_sha256": "%064d", "kind": "FORM", "ID": "%s",`+
			` "Owner": "forms", "Contract": "forms.v2"}]}`,
		registryPackageID, 1, registryComponentID,
	))
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		postgresJSONB, packagecontract.MaximumManifestBytes,
	)
	if err != nil {
		t.Fatal(err)
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(canonical)
	value, err := scanVersion(registryVersionRow{manifest: postgresJSONB,
		digest: hex.EncodeToString(digest[:]), signature: ed25519.Sign(privateKey, canonical)})
	if err != nil {
		t.Fatalf("scan signed version: %v", err)
	}
	manifest, err := packagecontract.Verify(value.Manifest, packagecontract.Attestation{
		ManifestSHA256: value.ManifestSHA256, Signature: value.Signature, KeyID: value.KeyID,
	}, publicKey)
	if err != nil || manifest.PackageID.String() != registryPackageID {
		t.Fatalf("verify scanned manifest = %s, %v", manifest.PackageID.String(), err)
	}
}

func TestResolveVersionFailsClosedForInvalidRegistryManifest(t *testing.T) {
	storage := &store{
		queryRow: func(context.Context, string, ...any) row {
			return registryVersionRow{manifest: []byte(`{"slug":`)}
		},
		query:    func(context.Context, string, ...any) (rows, error) { return nil, nil },
		notFound: func(error) bool { return false },
		allowed:  func(context.Context) bool { return true },
	}
	_, err := storage.ResolveVersion(context.Background(), registryPackageID, "1.0.0")
	if !errors.Is(err, packageregistry.ErrUnavailable) {
		t.Fatalf("resolve invalid manifest error = %v", err)
	}
}
