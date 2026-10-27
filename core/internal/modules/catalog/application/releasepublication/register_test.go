package releasepublication

import (
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
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

const publicationPackageID = "018f6000-0000-7000-8000-000000000001"
const publicationComponentID = "018f6000-0000-7000-8000-000000000002"
const publicationReleaseID = "018f6000-0000-7000-8000-000000000003"

type publicationRegistrar struct {
	values []packageregistry.SealedVersion
	err    error
}

func (registrar *publicationRegistrar) Register(
	_ context.Context,
	value packageregistry.SealedVersion,
) error {
	registrar.values = append(registrar.values, value.Clone())
	return registrar.err
}

type publicationKeys struct {
	key ed25519.PublicKey
	err error
}

func (keys publicationKeys) Resolve(string) (ed25519.PublicKey, error) {
	return append(ed25519.PublicKey(nil), keys.key...), keys.err
}

func TestRegisterVerifiesSignedCanonicalReleaseEvidence(t *testing.T) {
	value, publicKey := signedPublication(t)
	registrar := &publicationRegistrar{}
	application, err := NewApplication(registrar, publicationKeys{key: publicKey})
	if err != nil {
		t.Fatal(err)
	}
	if err = application.Register(context.Background(), value); err != nil {
		t.Fatalf("register: %v", err)
	}
	if len(registrar.values) != 1 || registrar.values[0].ReleaseID != publicationReleaseID {
		t.Fatalf("registered values = %#v", registrar.values)
	}
	value.Manifest[0] = '!'
	if registrar.values[0].Manifest[0] == '!' {
		t.Fatal("registrar input aliases caller manifest")
	}
}

func TestRegisterRejectsUntrustedOrMismatchedEvidenceBeforeStorage(t *testing.T) {
	valid, publicKey := signedPublication(t)
	tests := []struct {
		name   string
		mutate func(*packageregistry.SealedVersion)
		keys   publicationKeys
	}{
		{"noncanonical", func(value *packageregistry.SealedVersion) {
			value.Manifest = append([]byte(" \n"), value.Manifest...)
		}, publicationKeys{key: publicKey}},
		{"signature", func(value *packageregistry.SealedVersion) {
			value.Signature[0] ^= 0xff
		}, publicationKeys{key: publicKey}},
		{"identity", func(value *packageregistry.SealedVersion) {
			value.Slug = "other-package"
		}, publicationKeys{key: publicKey}},
		{"release", func(value *packageregistry.SealedVersion) {
			value.ReleaseID = "not-a-release"
		}, publicationKeys{key: publicKey}},
		{"timestamp", func(value *packageregistry.SealedVersion) {
			value.RegisteredAt = value.RegisteredAt.Add(time.Nanosecond)
		}, publicationKeys{key: publicKey}},
		{"owner payload", func(value *packageregistry.SealedVersion) {
			value.OwnerPayloads[0].Document[1] ^= 1
		}, publicationKeys{key: publicKey}},
		{"key", func(*packageregistry.SealedVersion) {},
			publicationKeys{err: errors.New("unknown key")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := valid.Clone()
			test.mutate(&value)
			registrar := &publicationRegistrar{}
			application, err := NewApplication(registrar, test.keys)
			if err != nil {
				t.Fatal(err)
			}
			err = application.Register(context.Background(), value)
			if !errors.Is(err, packageregistry.ErrInvalid) || len(registrar.values) != 0 {
				t.Fatalf("error = %v, registrations = %d", err, len(registrar.values))
			}
		})
	}
}

func signedPublication(t *testing.T) (packageregistry.SealedVersion, ed25519.PublicKey) {
	t.Helper()
	payloadDocument, err := canonicaljsonv1.CanonicalizeObject(
		[]byte(fmt.Sprintf(
			`{"component_id":"%s","definition":{"slug":"catalog-test"}}`,
			publicationComponentID,
		)),
		catalogpackagev1.MaximumPayloadBytes,
	)
	if err != nil {
		t.Fatal(err)
	}
	payloadDigest := sha256.Sum256(payloadDocument)
	payloadDigestHex := hex.EncodeToString(payloadDigest[:])
	document := []byte(fmt.Sprintf(
		`{"contract":"catalog-package.v1","package_id":"%s","slug":"catalog-test",`+
			`"version":"1.0.0","platform_contract":"harpy.v2","profile":"business",`+
			`"dependencies":[],"components":[{"ID":"%s","Owner":"iam",`+
			`"Contract":"iam.template.v1","kind":"IAM_TEMPLATE","content_sha256":"%s"}],"menu":[]}`,
		publicationPackageID, publicationComponentID, payloadDigestHex,
	))
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, packagecontract.MaximumManifestBytes,
	)
	if err != nil {
		t.Fatal(err)
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(canonical)
	return packageregistry.SealedVersion{
		PackageID: publicationPackageID, Slug: "catalog-test", Version: "1.0.0",
		Manifest: canonical, ManifestSHA256: hex.EncodeToString(digest[:]),
		Signature: ed25519.Sign(privateKey, canonical), KeyID: "release-2026-07",
		ReleaseID:    publicationReleaseID,
		RegisteredAt: time.Date(2026, 7, 30, 12, 0, 0, 123000, time.UTC),
		OwnerPayloads: []catalogpackagev1.OwnerPayload{{
			ComponentID: publicationComponentID, Owner: "iam",
			Contract: "iam.template.v1", ContentSHA256: payloadDigestHex,
			Document: payloadDocument,
		}},
	}, publicKey
}
