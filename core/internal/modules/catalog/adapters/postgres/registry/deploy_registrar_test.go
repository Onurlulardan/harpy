package registry

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

const deployPackageID = "018f7000-0000-7000-8000-000000000001"
const deployComponentID = "018f7000-0000-7000-8000-000000000002"
const deployReleaseID = "018f7000-0000-7000-8000-000000000003"

type deployVersionRow struct {
	value               packageregistry.SealedVersion
	payloadReleaseID    string
	payloadRegisteredAt time.Time
	err                 error
}

func (source deployVersionRow) Scan(destinations ...any) error {
	if source.err != nil {
		return source.err
	}
	if len(destinations) != 10 {
		return fmt.Errorf("deploy version destinations = %d", len(destinations))
	}
	*(destinations[0].(*string)) = source.value.PackageID
	*(destinations[1].(*string)) = source.value.Slug
	*(destinations[2].(*string)) = source.value.Version
	*(destinations[3].(*[]byte)) = append([]byte(nil), source.value.Manifest...)
	*(destinations[4].(*string)) = source.value.ManifestSHA256
	*(destinations[5].(*[]byte)) = append([]byte(nil), source.value.Signature...)
	*(destinations[6].(*string)) = source.value.KeyID
	*(destinations[7].(*string)) = source.value.ReleaseID
	*(destinations[8].(*time.Time)) = source.value.RegisteredAt
	releaseID := source.payloadReleaseID
	if releaseID == "" {
		releaseID = source.value.ReleaseID
	}
	registeredAt := source.payloadRegisteredAt
	if registeredAt.IsZero() {
		registeredAt = source.value.RegisteredAt
	}
	stored := make([]storedOwnerPayload, len(source.value.OwnerPayloads))
	for index := range source.value.OwnerPayloads {
		stored[index] = storedOwnerPayload{
			OwnerPayload: source.value.OwnerPayloads[index],
			ReleaseID:    releaseID,
			RegisteredAt: registeredAt,
		}
	}
	document, err := json.Marshal(stored)
	if err != nil {
		return err
	}
	payloads, err := canonicaljsonv1.Canonicalize(
		document, catalogpackagev1.MaximumPackageSetBytes,
	)
	if err != nil {
		return err
	}
	*(destinations[9].(*[]byte)) = payloads
	return nil
}

func TestDeployRegistrarRegistersAndAcceptsExactReplay(t *testing.T) {
	value := deployVersion(t)
	var statement string
	var arguments []any
	registrar := newDeployRegistrar(
		func(_ context.Context, query string, values ...any) DeployRow {
			statement, arguments = query, append([]any(nil), values...)
			return deployVersionRow{value: value}
		},
		func(err error) bool { return errors.Is(err, sql.ErrNoRows) },
	)
	if err := registrar.Register(context.Background(), value); err != nil {
		t.Fatalf("register exact evidence: %v", err)
	}
	if !strings.Contains(statement, "INSERT INTO platform.catalog_package_versions") ||
		!strings.Contains(statement, "INSERT INTO platform.catalog_package_owner_payloads") ||
		len(arguments) != 10 || arguments[3] != string(value.Manifest) {
		t.Fatalf("query arguments = %#v", arguments)
	}
}

func TestDeployRegistrarAcceptsExactImmutableVersionAcrossReleases(t *testing.T) {
	stored := deployVersion(t)
	replayed := stored.Clone()
	replayed.ReleaseID = "018f7000-0000-7000-8000-000000000004"
	replayed.RegisteredAt = replayed.RegisteredAt.Add(time.Hour)
	registrar := newDeployRegistrar(
		func(context.Context, string, ...any) DeployRow {
			return deployVersionRow{value: stored}
		},
		func(error) bool { return false },
	)
	if err := registrar.Register(context.Background(), replayed); err != nil {
		t.Fatalf("cross-release exact replay error = %v", err)
	}
}

func TestDeployRegistrarRejectsCrossReleaseContentConflict(t *testing.T) {
	stored := deployVersion(t)
	replayed := deployVersion(t)
	stored.Manifest = []byte(`{}`)
	replayed.ReleaseID = "018f7000-0000-7000-8000-000000000004"
	replayed.RegisteredAt = replayed.RegisteredAt.Add(time.Hour)
	registrar := newDeployRegistrar(
		func(context.Context, string, ...any) DeployRow {
			return deployVersionRow{value: stored}
		},
		func(error) bool { return false },
	)
	if err := registrar.Register(context.Background(), replayed); !errors.Is(
		err, packageregistry.ErrConflict,
	) {
		t.Fatalf("cross-release content conflict error = %v", err)
	}
}

func TestDeployRegistrarRejectsImmutableConflictAndStorageFailure(t *testing.T) {
	value := deployVersion(t)
	conflicting := value.Clone()
	conflicting.Signature[0] ^= 0xff
	payloadConflict := value.Clone()
	payloadConflict.OwnerPayloads[0].Document = []byte(fmt.Sprintf(
		`{"component_id":"%s","definition":{"slug":"other"}}`,
		deployComponentID,
	))
	tests := []struct {
		name string
		row  row
		want error
	}{
		{"stored mismatch", deployVersionRow{value: conflicting}, packageregistry.ErrConflict},
		{"stored payload mismatch", deployVersionRow{value: payloadConflict},
			packageregistry.ErrConflict},
		{"stored payload provenance mismatch", deployVersionRow{
			value:            value,
			payloadReleaseID: "018f7000-0000-7000-8000-000000000004",
		}, packageregistry.ErrUnavailable},
		{"missing identity", deployVersionRow{err: sql.ErrNoRows}, packageregistry.ErrConflict},
		{"database error", deployVersionRow{err: errors.New("database down")},
			packageregistry.ErrUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registrar := newDeployRegistrar(
				func(context.Context, string, ...any) DeployRow { return test.row },
				func(err error) bool { return errors.Is(err, sql.ErrNoRows) },
			)
			if err := registrar.Register(context.Background(), value); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestDeployRegistrarRejectsInvalidEvidenceBeforeQuery(t *testing.T) {
	value := deployVersion(t)
	value.ManifestSHA256 = strings.Repeat("f", 64)
	called := false
	registrar := newDeployRegistrar(
		func(context.Context, string, ...any) DeployRow {
			called = true
			return deployVersionRow{}
		},
		func(error) bool { return false },
	)
	err := registrar.Register(context.Background(), value)
	if !errors.Is(err, packageregistry.ErrInvalid) || called {
		t.Fatalf("error = %v, query called = %t", err, called)
	}
}

func TestNewDeployRegistrarRequiresDatabasePort(t *testing.T) {
	if _, err := NewDeployRegistrar(nil, nil); !errors.Is(err, packageregistry.ErrUnavailable) {
		t.Fatalf("constructor error = %v", err)
	}
}
