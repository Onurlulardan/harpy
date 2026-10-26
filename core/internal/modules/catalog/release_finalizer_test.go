package catalog

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type catalogFinalizerKeys struct {
	key ed25519.PublicKey
}

func (keys catalogFinalizerKeys) Resolve(keyID string) (ed25519.PublicKey, error) {
	if keyID != "catalog-2026-07" || len(keys.key) != ed25519.PublicKeySize {
		return nil, errors.New("Catalog package key is unavailable")
	}
	return append(ed25519.PublicKey(nil), keys.key...), nil
}

type catalogFinalizerRow struct {
	arguments []any
}

func (row catalogFinalizerRow) Scan(destinations ...any) error {
	if len(destinations) != 10 || len(row.arguments) != 10 {
		return errors.New("unexpected Catalog registration row")
	}
	*destinations[0].(*string) = row.arguments[0].(string)
	*destinations[1].(*string) = row.arguments[1].(string)
	*destinations[2].(*string) = row.arguments[2].(string)
	*destinations[3].(*[]byte) = []byte(row.arguments[3].(string))
	*destinations[4].(*string) = row.arguments[4].(string)
	*destinations[5].(*[]byte) = append([]byte(nil), row.arguments[5].([]byte)...)
	*destinations[6].(*string) = row.arguments[6].(string)
	*destinations[7].(*string) = row.arguments[7].(string)
	*destinations[8].(*time.Time) = row.arguments[8].(time.Time)
	var payloads []catalogpackagev1.OwnerPayload
	if err := json.Unmarshal([]byte(row.arguments[9].(string)), &payloads); err != nil {
		return err
	}
	stored := make([]map[string]any, len(payloads))
	for index := range payloads {
		stored[index] = map[string]any{
			"component_id":   payloads[index].ComponentID,
			"owner":          payloads[index].Owner,
			"contract":       payloads[index].Contract,
			"content_sha256": payloads[index].ContentSHA256,
			"document":       payloads[index].Document,
			"release_id":     row.arguments[7].(string),
			"registered_at":  row.arguments[8].(time.Time),
		}
	}
	document, err := json.Marshal(stored)
	if err != nil {
		return err
	}
	canonical, err := canonicaljsonv1.Canonicalize(
		document, catalogpackagev1.MaximumPackageSetBytes,
	)
	if err != nil {
		return err
	}
	*destinations[9].(*[]byte) = canonical
	return nil
}

func TestCatalogReleaseFinalizerUsesVerifiedReleaseEvidence(t *testing.T) {
	release, version, key := catalogFinalizerEvidence(t)
	finalizer, err := NewReleaseFinalizer(
		release, catalogFinalizerKeys{key: key},
		[]packageregistry.SealedVersion{version},
	)
	if err != nil {
		t.Fatal(err)
	}
	var captured []any
	queryRow := func(
		_ context.Context,
		statement string,
		arguments ...any,
	) platformmigrator.FinalizationRow {
		if statement == "" {
			t.Fatal("Catalog registration SQL is empty")
		}
		captured = append([]any(nil), arguments...)
		return catalogFinalizerRow{arguments: arguments}
	}
	if err := finalizer.FinalizeRelease(context.Background(), queryRow); err != nil {
		t.Fatalf("finalize Catalog release: %v", err)
	}
	if len(captured) != 10 ||
		captured[7] != catalogFinalizerReleaseID ||
		captured[8] != catalogFinalizerRegisteredAt {
		t.Fatalf("release-derived registration = %#v", captured)
	}
}

func TestCatalogReleaseFinalizerRejectsForeignPackageSigner(t *testing.T) {
	release, version, key := catalogFinalizerEvidence(t)
	version.KeyID = "foreign-release"
	finalizer, err := NewReleaseFinalizer(
		release, catalogFinalizerKeys{key: key},
		[]packageregistry.SealedVersion{version},
	)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	queryRow := func(
		context.Context,
		string,
		...any,
	) platformmigrator.FinalizationRow {
		called = true
		return nil
	}
	err = finalizer.FinalizeRelease(context.Background(), queryRow)
	if !errors.Is(err, packageregistry.ErrInvalid) || called {
		t.Fatalf("foreign signer result = %v, queried=%t", err, called)
	}
}

func TestCatalogReleaseFinalizerRequiresPackageEvidence(t *testing.T) {
	if _, err := NewReleaseFinalizer(
		attestation.VerifiedRelease{}, nil, nil,
	); err == nil {
		t.Fatal("incomplete Catalog release evidence was accepted")
	}
}
