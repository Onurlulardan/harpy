package iam

import (
	"context"
	"errors"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	querycatalog "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type importCatalogReaderFixture struct {
	schema profileapp.Schema
	err    error
	calls  int
	kind   profileapp.ResourceKind
}

func (reader *importCatalogReaderFixture) ActiveSchema(
	_ context.Context,
	kind profileapp.ResourceKind,
) (profileapp.Schema, error) {
	reader.calls++
	reader.kind = kind
	return reader.schema, reader.err
}

func TestSystemTransferImportCatalogPinAcceptsExactActiveCatalog(t *testing.T) {
	schema := importCatalogSchema(t, "019be300-0000-7000-8000-000000000001", 1)
	attempt := importCatalogPinAttempt(t, schema)
	reader := &importCatalogReaderFixture{schema: schema}
	if err := verifySystemTransferImportCatalogPinWithReader(
		t.Context(), reader, attempt,
	); err != nil {
		t.Fatal(err)
	}
	if reader.calls != 1 || reader.kind != profileapp.ResourceUser {
		t.Fatalf("calls=%d kind=%s", reader.calls, reader.kind)
	}
}

func TestSystemTransferImportCatalogPinRejectsDrift(t *testing.T) {
	schema := importCatalogSchema(t, "019be300-0000-7000-8000-000000000002", 1)
	attempt := importCatalogPinAttempt(t, schema)
	tests := []struct {
		name   string
		reader *importCatalogReaderFixture
		mutate func(*systemtransfer.Attempt)
		want   error
	}{
		{name: "catalog digest", reader: &importCatalogReaderFixture{schema: schema},
			mutate: func(value *systemtransfer.Attempt) { value.CatalogSHA256[0] ^= 0xff },
			want:   systemtransfer.ErrImportCatalogPinStale},
		{name: "identity manifest", reader: &importCatalogReaderFixture{schema: schema},
			mutate: func(value *systemtransfer.Attempt) {
				value.IdentityManifestSHA256[0] ^= 0xff
			}, want: systemtransfer.ErrImportCatalogPinStale},
		{name: "active version", reader: &importCatalogReaderFixture{schema: importCatalogSchema(
			t, "019be300-0000-7000-8000-000000000003", 2,
		)}, mutate: func(*systemtransfer.Attempt) {},
			want: systemtransfer.ErrImportCatalogPinStale},
		{name: "reader failure", reader: &importCatalogReaderFixture{
			err: errors.New("catalog unavailable"),
		}, mutate: func(*systemtransfer.Attempt) {},
			want: systemtransfer.ErrWorkerUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := attempt
			test.mutate(&candidate)
			if err := verifySystemTransferImportCatalogPinWithReader(
				t.Context(), test.reader, candidate,
			); !errors.Is(err, test.want) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func importCatalogSchema(
	t *testing.T,
	schemaID string,
	version int64,
) profileapp.Schema {
	t.Helper()
	displayName, err := profileapp.NewField(profileapp.FieldInput{
		ID:       uuid.MustParse("e1000000-0000-4000-8000-0000000000ff"),
		Resource: profileapp.ResourceUser, Key: "display_name",
		Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
		Required: true, Classification: profileapp.ClassificationPublic,
		Normalization: profileapp.NormalizationNoneV1, State: profileapp.FieldActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	return profilePublicationSchemaFields(t, schemaID, version, displayName)
}

func importCatalogPinAttempt(
	t *testing.T,
	schema profileapp.Schema,
) systemtransfer.Attempt {
	t.Helper()
	custom, err := profileapp.NewQueryCatalog(schema)
	if err != nil {
		t.Fatal(err)
	}
	published, err := querycatalog.Build(custom)
	if err != nil {
		t.Fatal(err)
	}
	pin, err := systemTransferPublishedPin(published)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := systemtransfer.ManifestFor(publicresource.Users)
	if err != nil {
		t.Fatal(err)
	}
	return systemtransfer.Attempt{
		Resource: publicresource.Users, CatalogVersionID: pin.versionID,
		CatalogSHA256: pin.sha256, IdentityManifestSHA256: manifest.Digest,
	}
}
