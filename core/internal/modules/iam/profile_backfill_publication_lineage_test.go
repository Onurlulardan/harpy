package iam

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func TestProfileEvolutionRejectsSilentActiveFieldOmission(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, fieldID, false, 128)
	target := profilePublicationSchemaFields(t,
		"13000000-0000-4000-8000-000000000002", 2)
	if _, err := profileBackfillRequired(source, target, nil); !errors.Is(
		err, ErrProfileBackfillUnmappable,
	) {
		t.Fatalf("silent omission error = %v", err)
	}
}

func TestProfileReplacementRequiresRetiredSourceRevision(t *testing.T) {
	sourceID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	targetID := uuid.MustParse("12000000-0000-4000-8000-000000000002")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, sourceID, false, 128)
	target := profilePublicationSchemaFields(t,
		"13000000-0000-4000-8000-000000000002", 2,
		profilePublicationField(t, sourceID, "display_name", false, 128,
			profileapp.FieldActive),
		profilePublicationField(t, targetID, "display_name_v2", true, 128,
			profileapp.FieldActive),
	)
	mapping := []ProfileMigrationMappingInput{{
		SourceFieldID: sourceID, TargetFieldID: targetID,
		ConverterKind: profilebackfillapp.MappingNormalize,
	}}
	if _, err := profileBackfillRequired(source, target, mapping); !errors.Is(
		err, ErrProfileBackfillUnmappable,
	) {
		t.Fatalf("non-retiring replacement error = %v", err)
	}
}

func TestProfileEvolutionRejectsRetiredFieldReactivation(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	source := profilePublicationSchemaFields(t,
		"13000000-0000-4000-8000-000000000001", 1,
		profilePublicationField(t, fieldID, "display_name", false, 128,
			profileapp.FieldRetired),
	)
	target := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 128)
	if _, err := profileBackfillRequired(source, target, nil); !errors.Is(
		err, ErrProfileBackfillUnmappable,
	) {
		t.Fatalf("reactivation error = %v", err)
	}
}

func TestProfilePageSourceEvidencePinsActualSchemaVersion(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, fieldID, false, 128)
	target := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 64)
	mappings, err := profileBackfillMappings(source, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	shared, err := profileBackfillMigrationMappings(mappings)
	if err != nil {
		t.Fatal(err)
	}
	resource := profilebackfillapp.SourceResource{
		ID:              uuid.MustParse("14000000-0000-4000-8000-000000000001"),
		SchemaVersionID: source.ID(), Revision: 1,
		Document: []byte("{}"), SHA256: sha256.Sum256([]byte("{}")),
	}
	pageOne, err := profileBackfillPage(
		uuid.MustParse("15000000-0000-4000-8000-000000000001"), 1,
		time.Unix(1, 0).UTC(), uuid.Nil, source, target, shared,
		map[uuid.UUID]profileapp.Schema{source.ID(): source},
		[]profilebackfillapp.SourceResource{resource},
	)
	if err != nil {
		t.Fatal(err)
	}
	alternate := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000009", 1, fieldID, false, 128)
	resource.SchemaVersionID = alternate.ID()
	pageTwo, err := profileBackfillPage(
		uuid.MustParse("15000000-0000-4000-8000-000000000001"), 1,
		time.Unix(1, 0).UTC(), uuid.Nil, source, target, shared,
		map[uuid.UUID]profileapp.Schema{alternate.ID(): alternate},
		[]profilebackfillapp.SourceResource{resource},
	)
	if err != nil || pageOne.SourceSHA256 == pageTwo.SourceSHA256 {
		t.Fatalf("actual schema lineage was not pinned: %v", err)
	}
}
