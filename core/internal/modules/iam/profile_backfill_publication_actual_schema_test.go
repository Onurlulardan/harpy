package iam

import (
	"crypto/sha256"
	"testing"
	"time"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func TestProfileBackfillPageUsesActualHistoricalValidationContract(t *testing.T) {
	oldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	optionalID := uuid.MustParse("12000000-0000-4000-8000-000000000002")
	replacementID := uuid.MustParse("12000000-0000-4000-8000-000000000003")
	actual := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, oldID, false, 64)
	current := profilePublicationSchemaFields(t,
		"13000000-0000-4000-8000-000000000002", 2,
		profilePublicationField(t, oldID, "display_name", false, 8, profileapp.FieldActive),
		profilePublicationField(t, optionalID, "nickname", false, 8, profileapp.FieldActive),
	)
	target := profilePublicationSchemaFields(t,
		"13000000-0000-4000-8000-000000000003", 3,
		profilePublicationField(t, oldID, "display_name", false, 64, profileapp.FieldRetired),
		profilePublicationField(t, optionalID, "nickname", false, 8, profileapp.FieldActive),
		profilePublicationField(t, replacementID, "contact_alias_v2", true, 128, profileapp.FieldActive),
	)
	mappings, err := profileBackfillMappings(current, target, []ProfileMigrationMappingInput{{
		SourceFieldID: oldID, TargetFieldID: replacementID,
		ConverterKind: profilebackfillapp.MappingNormalize,
	}})
	if err != nil {
		t.Fatal(err)
	}
	shared, err := profileBackfillMigrationMappings(mappings)
	if err != nil {
		t.Fatal(err)
	}
	document := []byte(`{"12000000-0000-4000-8000-000000000001":"historical-value"}`)
	resource := profilebackfillapp.SourceResource{
		ID:              uuid.MustParse("14000000-0000-4000-8000-000000000001"),
		SchemaVersionID: actual.ID(), Revision: 1, Document: document,
		SHA256: sha256.Sum256(document),
	}
	page, err := profileBackfillPage(
		uuid.MustParse("15000000-0000-4000-8000-000000000001"), 1,
		time.Unix(1, 0).UTC(), uuid.Nil, current, target, shared,
		map[uuid.UUID]profileapp.Schema{actual.ID(): actual},
		[]profilebackfillapp.SourceResource{resource},
	)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := profileapp.SelectMigrationMappingsForSource(actual, target, shared)
	if err != nil {
		t.Fatal(err)
	}
	converted, err := profileBackfillDocument(actual, target, selected, resource)
	if err != nil {
		t.Fatal(err)
	}
	digest := converted.SHA256()
	expected := sha256.New()
	writeProfileBackfillDigest(expected, []byte(resource.ID.String()), digest[:])
	var expectedSHA [sha256.Size]byte
	copy(expectedSHA[:], expected.Sum(nil))
	if page.ExpectedTargetSHA256 != expectedSHA {
		t.Fatalf("target digest = %x, want %x", page.ExpectedTargetSHA256, expectedSHA)
	}
}
