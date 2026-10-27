//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	registryadapter "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/registry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	releasepublication "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/releasepublication"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

const publicationIntegrationPackageID = "018f8000-0000-7000-8000-000000000001"
const publicationIntegrationComponentID = "018f8000-0000-7000-8000-000000000002"
const publicationIntegrationReleaseID = "018f8000-0000-7000-8000-000000000003"

type publicationIntegrationKeys struct{ key ed25519.PublicKey }

func (keys publicationIntegrationKeys) Resolve(string) (ed25519.PublicKey, error) {
	return append(ed25519.PublicKey(nil), keys.key...), nil
}

func assertCatalogRegistryPublication(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	value, publicKey := catalogPublicationEvidence(t)
	registrar, err := registryadapter.NewDeployRegistrar(
		func(callCtx context.Context, statement string, arguments ...any) registryadapter.DeployRow {
			return database.QueryRowContext(callCtx, statement, arguments...)
		},
		func(err error) bool { return errors.Is(err, sql.ErrNoRows) },
	)
	if err != nil {
		t.Fatal(err)
	}
	application, err := releasepublication.NewApplication(
		registrar, publicationIntegrationKeys{key: publicKey},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err = application.Register(ctx, value); err != nil {
		t.Fatalf("publish signed Catalog package: %v", err)
	}
	if err = application.Register(ctx, value); err != nil {
		t.Fatalf("replay signed Catalog package: %v", err)
	}
	crossRelease := value.Clone()
	crossRelease.ReleaseID = "018f8000-0000-7000-8000-000000000004"
	crossRelease.RegisteredAt = crossRelease.RegisteredAt.Add(time.Hour)
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	transactionRegistrar, err := registryadapter.NewDeployRegistrar(
		func(
			callCtx context.Context,
			statement string,
			arguments ...any,
		) registryadapter.DeployRow {
			return transaction.QueryRowContext(callCtx, statement, arguments...)
		},
		func(err error) bool { return errors.Is(err, sql.ErrNoRows) },
	)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	transactionApplication, err := releasepublication.NewApplication(
		transactionRegistrar, publicationIntegrationKeys{key: publicKey},
	)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	if err = transactionApplication.Register(ctx, crossRelease); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("cross-release exact Catalog package replay error = %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit cross-release exact replay: %v", err)
	}
	conflicting := value.Clone()
	conflicting.KeyID = "catalog-key-rotated"
	if err = application.Register(ctx, conflicting); !errors.Is(err, packageregistry.ErrConflict) {
		t.Fatalf("immutable publication conflict = %v", err)
	}
	var packageCount, versionCount, payloadCount int
	var releaseID, digest, componentID, owner, contract, contentSHA256 string
	var payloadDocument []byte
	err = database.QueryRowContext(ctx, `SELECT
 (SELECT count(*) FROM platform.catalog_packages WHERE package_id=$1),
 (SELECT count(*) FROM platform.catalog_package_versions
  WHERE package_id=$1 AND version=$2),
 (SELECT count(*) FROM platform.catalog_package_owner_payloads
  WHERE package_id=$1 AND version=$2),
 version.release_id::text, version.manifest_sha256,
 payload.component_id::text, payload.owner, payload.contract,
 payload.content_sha256, payload.document
FROM platform.catalog_package_versions AS version
JOIN platform.catalog_package_owner_payloads AS payload
  ON payload.package_id=version.package_id AND payload.version=version.version
WHERE version.package_id=$1 AND version.version=$2`,
		value.PackageID, value.Version).
		Scan(&packageCount, &versionCount, &payloadCount, &releaseID, &digest,
			&componentID, &owner, &contract, &contentSHA256, &payloadDocument)
	canonicalPayload, canonicalErr := canonicaljsonv1.CanonicalizeObject(
		payloadDocument, catalogpackagev1.MaximumPayloadBytes,
	)
	if err != nil || packageCount != 1 || versionCount != 1 || payloadCount != 1 ||
		releaseID != value.ReleaseID || digest != value.ManifestSHA256 ||
		componentID != publicationIntegrationComponentID || owner != "forms" ||
		contract != "forms.v2" ||
		contentSHA256 != value.OwnerPayloads[0].ContentSHA256 ||
		canonicalErr != nil ||
		!bytes.Equal(canonicalPayload, value.OwnerPayloads[0].Document) {
		t.Fatalf("published registry evidence = %d/%d/%d %s %s %s %s %s %v err=%v",
			packageCount, versionCount, payloadCount, releaseID, digest,
			componentID, owner, contract, payloadDocument, err)
	}
}

func catalogPublicationEvidence(
	t *testing.T,
) (packageregistry.SealedVersion, ed25519.PublicKey) {
	t.Helper()
	payloadDocument, err := canonicaljsonv1.CanonicalizeObject(
		[]byte(fmt.Sprintf(
			`{"component_id":"%s","definition":{"slug":"publication-proof"}}`,
			publicationIntegrationComponentID,
		)),
		catalogpackagev1.MaximumPayloadBytes,
	)
	if err != nil {
		t.Fatal(err)
	}
	payloadDigest := sha256.Sum256(payloadDocument)
	payloadDigestHex := hex.EncodeToString(payloadDigest[:])
	document := []byte(fmt.Sprintf(
		`{"contract":"catalog-package.v1","package_id":"%s","slug":"publication-proof",`+
			`"version":"1.0.0","platform_contract":"harpy.v2","profile":"business",`+
			`"dependencies":[],"components":[{"ID":"%s","Owner":"forms",`+
			`"Contract":"forms.v2","kind":"FORM","content_sha256":"%s"}],"menu":[]}`,
		publicationIntegrationPackageID, publicationIntegrationComponentID,
		payloadDigestHex,
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
		PackageID: publicationIntegrationPackageID, Slug: "publication-proof",
		Version: "1.0.0", Manifest: canonical,
		ManifestSHA256: hex.EncodeToString(digest[:]),
		Signature:      ed25519.Sign(privateKey, canonical), KeyID: "release-2026-07",
		ReleaseID:    publicationIntegrationReleaseID,
		RegisteredAt: time.Date(2026, 7, 30, 12, 0, 0, 123000, time.UTC),
		OwnerPayloads: []catalogpackagev1.OwnerPayload{{
			ComponentID: publicationIntegrationComponentID, Owner: "forms",
			Contract: "forms.v2", ContentSHA256: payloadDigestHex,
			Document: payloadDocument,
		}},
	}, publicKey
}
