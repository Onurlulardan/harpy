package catalog

import (
	"slices"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

func TestBuildPublishesSafeCoreSearchFields(t *testing.T) {
	tests := []struct {
		kind profile.ResourceKind
		want []string
	}{
		{profile.ResourceRole, []string{"name", "description"}},
		{profile.ResourceOrganization, []string{"name", "slug"}},
	}
	for _, test := range tests {
		t.Run(string(test.kind), func(t *testing.T) {
			published := buildFixture(t, test.kind, nil)
			fieldCatalog := rootCatalog(t, published)
			search, err := fieldCatalog.SearchFields()
			if err != nil || len(search) != len(test.want) {
				t.Fatalf("search fields = %v, error = %v", search, err)
			}
			for index, identifier := range search {
				assertSearchField(t, fieldCatalog, identifier, test.want[index])
			}
		})
	}
}

func TestBuildSearchesUsersOnlyByPublicProfileDisplayName(t *testing.T) {
	displayID := uuid.MustParse("e1000000-0000-4000-8000-000000000010")
	display, err := profile.NewField(profile.FieldInput{
		ID: displayID, Resource: profile.ResourceUser, Key: "display_name",
		Type: profile.LogicalString, Cardinality: profile.CardinalitySingle,
		Classification: profile.ClassificationPublic,
		Normalization:  profile.NormalizationNoneV1, State: profile.FieldActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	fieldCatalog := rootCatalog(
		t, buildFixture(t, profile.ResourceUser, []profile.Field{display}),
	)
	search, err := fieldCatalog.SearchFields()
	if err != nil || len(search) != 1 || search[0] != displayID {
		t.Fatalf("user search fields = %v, error = %v", search, err)
	}
	for _, confidential := range []uuid.UUID{
		uuid.MustParse(userEmailFieldID),
		uuid.MustParse(userPortalIdentifierFieldID),
	} {
		if slices.Contains(search, confidential) {
			t.Fatal("confidential identifier leaked into user search")
		}
	}
	assertSearchField(t, fieldCatalog, displayID, "display_name")
}

func TestBuildRejectsNonPublicUserDisplayNameSearch(t *testing.T) {
	display, err := profile.NewField(profile.FieldInput{
		ID:       uuid.MustParse("e1000000-0000-4000-8000-000000000011"),
		Resource: profile.ResourceUser, Key: "display_name",
		Type: profile.LogicalString, Cardinality: profile.CardinalitySingle,
		Classification: profile.ClassificationInternal,
		Normalization:  profile.NormalizationNoneV1, State: profile.FieldActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	schema, err := profile.NewPublishedSchema(profile.SchemaInput{
		ID:       uuid.MustParse("f1000000-0000-4000-8000-000000000011"),
		Resource: profile.ResourceUser, Version: 1,
		State: profile.SchemaActive, Fields: []profile.Field{display},
	})
	if err != nil {
		t.Fatal(err)
	}
	queryCatalog, _ := profile.NewQueryCatalog(schema)
	if _, err := Build(queryCatalog); err == nil {
		t.Fatal("non-public user display_name unexpectedly became searchable")
	}
}

func TestBuildRejectsUserCatalogWithoutPublicDisplayName(t *testing.T) {
	schema, err := profile.NewPublishedSchema(profile.SchemaInput{
		ID:       uuid.MustParse("f1000000-0000-4000-8000-000000000012"),
		Resource: profile.ResourceUser, Version: 1,
		State: profile.SchemaActive, Fields: []profile.Field{},
	})
	if err != nil {
		t.Fatal(err)
	}
	queryCatalog, _ := profile.NewQueryCatalog(schema)
	if _, err := Build(queryCatalog); err == nil {
		t.Fatal("user catalog without public display_name unexpectedly published")
	}
}

func rootCatalog(t *testing.T, published Published) api.FieldCatalog {
	t.Helper()
	set, err := published.CatalogSet()
	if err != nil {
		t.Fatal(err)
	}
	fieldCatalog, err := set.RootCatalog()
	if err != nil {
		t.Fatal(err)
	}
	return fieldCatalog
}

func assertSearchField(
	t *testing.T,
	fieldCatalog api.FieldCatalog,
	identifier uuid.UUID,
	wantName string,
) {
	t.Helper()
	field, found, err := fieldCatalog.Field(identifier)
	if err != nil || !found {
		t.Fatalf("search field %s is unavailable", identifier)
	}
	name, _ := field.PublicName()
	capabilities, _ := field.Capabilities()
	searchable, _ := capabilities.Searchable()
	storage, _ := field.Storage()
	folded, present, foldedErr := storage.FoldedCompanion()
	if name != wantName || !searchable || foldedErr != nil || !present {
		t.Fatalf("search field %s contract is incomplete", wantName)
	}
	expected, _ := sqlidentifier.MintStatic(wantName + "_folded")
	actualFingerprint, _ := sqlidentifier.Fingerprint(folded)
	expectedFingerprint, _ := sqlidentifier.Fingerprint(expected)
	if actualFingerprint != expectedFingerprint {
		t.Fatalf("search field %s folded storage mismatch", wantName)
	}
}
