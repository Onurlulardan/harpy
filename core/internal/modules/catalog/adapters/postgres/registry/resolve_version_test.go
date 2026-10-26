package registry

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

func TestResolveVersionRejectsInvalidVersion(t *testing.T) {
	_, err := (&store{}).ResolveVersion(context.Background(),
		"018f5000-0000-7000-8000-000000000001", "latest")
	if !errors.Is(err, packageregistry.ErrInvalid) {
		t.Fatalf("resolve version error = %v", err)
	}
}

func TestVersionPayloadsReturnsReleaseBoundCanonicalDocuments(t *testing.T) {
	expected := deployVersion(t)
	var statement string
	storage := &store{
		queryRow: func(_ context.Context, query string, _ ...any) row {
			statement = query
			return deployVersionRow{value: expected}
		},
		query: func(context.Context, string, ...any) (rows, error) {
			return nil, nil
		},
		notFound: func(error) bool { return false },
		allowed:  func(context.Context) bool { return true },
	}
	value, err := storage.VersionPayloads(
		context.Background(), expected.PackageID, expected.Version,
	)
	if err != nil {
		t.Fatalf("resolve version payloads: %v", err)
	}
	if !strings.Contains(statement, "catalog_package_owner_payloads") ||
		!sameVersion(value, expected) {
		t.Fatalf("resolved payload version = %#v", value)
	}
	value.OwnerPayloads[0].Document[0] ^= 1
	if string(expected.OwnerPayloads[0].Document) ==
		string(value.OwnerPayloads[0].Document) {
		t.Fatal("payload reader aliased stored evidence")
	}
}
