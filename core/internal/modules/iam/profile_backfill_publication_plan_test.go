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

func TestProfileBackfillPageIsDeterministicAndValid(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, fieldID, false, 128)
	target := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 128)
	mappings, err := profileBackfillMappings(source, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	shared, err := profileBackfillMigrationMappings(mappings)
	if err != nil {
		t.Fatal(err)
	}
	document := []byte(`{"12000000-0000-4000-8000-000000000001":"Alice"}`)
	resourceID := uuid.MustParse("14000000-0000-4000-8000-000000000001")
	planID := uuid.MustParse("15000000-0000-4000-8000-000000000001")
	page, err := profileBackfillPage(
		planID, 1, time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC), uuid.Nil,
		source, target, shared, map[uuid.UUID]profileapp.Schema{source.ID(): source},
		[]profilebackfillapp.SourceResource{{
			ID: resourceID, SchemaVersionID: source.ID(), Revision: 1, Document: document,
			SHA256: sha256.Sum256(document),
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if page.JobID.Version() != uuid.Version(7) ||
		page.JobID != profileBackfillJobID(planID, 1) || page.ExpectedCount != 1 {
		t.Fatalf("unexpected deterministic page: %+v", page)
	}
	definition := target.DefinitionSHA256()
	plan := profilebackfillapp.Plan{
		ID: planID, SourceSchemaID: source.ID(),
		Target: profilebackfillapp.TargetSchema{
			ID: target.ID(), ResourceKind: "user", Version: 2,
			DefinitionJSON: target.DefinitionJSON(), DefinitionSHA256: definition,
			Fields: profileBackfillTargetFields(target),
		},
		ConverterVersion: profilebackfillapp.ConverterTypedMap,
		SourceCount:      1, PageSize: profilebackfillapp.MaximumPageSize,
		InitialRollingSHA256: sha256.Sum256(nil), Mappings: mappings,
		Pages: []profilebackfillapp.Page{page},
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("built plan rejected: %v", err)
	}
}

func TestProfileBackfillMappingRejectsNewRequiredField(t *testing.T) {
	sourceID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	targetID := uuid.MustParse("12000000-0000-4000-8000-000000000002")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, sourceID, false, 128)
	target := profilePublicationSchemaFields(t,
		"13000000-0000-4000-8000-000000000002", 2,
		profilePublicationField(t, sourceID, "display_name", false, 128,
			profileapp.FieldRetired),
		profilePublicationField(t, targetID, "display_name_v2", true, 128,
			profileapp.FieldActive),
	)
	if _, err := profileBackfillMappings(source, target, nil); !errors.Is(
		err, ErrProfileBackfillUnmappable,
	) {
		t.Fatalf("mapping error = %v", err)
	}
}

func TestProfileBackfillRequirementDistinguishesOptionalAndMappedRequired(t *testing.T) {
	sourceID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	targetID := uuid.MustParse("12000000-0000-4000-8000-000000000002")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, sourceID, false, 128)
	optional := profilePublicationSchemaFields(t,
		"13000000-0000-4000-8000-000000000002", 2,
		profilePublicationField(t, sourceID, "display_name", false, 128,
			profileapp.FieldActive),
		profilePublicationField(t, targetID, "nickname", false, 128,
			profileapp.FieldActive),
	)
	if required, err := profileBackfillRequired(source, optional, nil); err != nil || required {
		t.Fatalf("optional addition requirement = %t/%v", required, err)
	}
	requiredTarget := profilePublicationSchemaFields(t,
		"13000000-0000-4000-8000-000000000003", 2,
		profilePublicationField(t, sourceID, "display_name", false, 128,
			profileapp.FieldRetired),
		profilePublicationField(t, targetID, "display_name_v2", true, 128,
			profileapp.FieldActive),
	)
	if _, err := profileBackfillRequired(source, requiredTarget, nil); !errors.Is(
		err, ErrProfileBackfillUnmappable,
	) {
		t.Fatalf("missing required converter error = %v", err)
	}
	migration := []ProfileMigrationMappingInput{{
		SourceFieldID: sourceID, TargetFieldID: targetID,
		ConverterKind: profilebackfillapp.MappingNormalize,
	}}
	if required, err := profileBackfillRequired(
		source, requiredTarget, migration,
	); err != nil || !required {
		t.Fatalf("mapped required evolution = %t/%v", required, err)
	}
}

func TestProfileBackfillRequirementPlansValidationChange(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, fieldID, false, 128)
	target := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 64)
	if required, err := profileBackfillRequired(source, target, nil); err != nil || !required {
		t.Fatalf("validation change requirement = %t/%v", required, err)
	}
}

func TestProfileBackfillPageRejectsInvalidRowBeforePlanMaterialization(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	resourceID := uuid.MustParse("14000000-0000-4000-8000-000000000001")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, fieldID, false, 128)
	target := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 3)
	mappings, err := profileBackfillMappings(source, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	shared, err := profileBackfillMigrationMappings(mappings)
	if err != nil {
		t.Fatal(err)
	}
	document := []byte(`{"12000000-0000-4000-8000-000000000001":"Alice"}`)
	sourceSHA := sha256.Sum256(document)
	_, err = profileBackfillPage(
		uuid.MustParse("15000000-0000-4000-8000-000000000001"), 1,
		time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC), uuid.Nil, source, target,
		shared, map[uuid.UUID]profileapp.Schema{source.ID(): source},
		[]profilebackfillapp.SourceResource{{
			ID: resourceID, SchemaVersionID: source.ID(), Revision: 1,
			Document: document, SHA256: sourceSHA,
		}},
	)
	if !errors.Is(err, ErrProfileBackfillUnmappable) {
		t.Fatalf("invalid source row error = %v", err)
	}
}
