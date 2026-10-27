package iam

import (
	"encoding/json"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func TestProfileMigrationMappingOrderDoesNotChangeRequestHash(t *testing.T) {
	sourceA := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	sourceB := uuid.MustParse("12000000-0000-4000-8000-000000000002")
	targetA := uuid.MustParse("13000000-0000-4000-8000-000000000001")
	targetB := uuid.MustParse("13000000-0000-4000-8000-000000000002")
	first := []ProfileMigrationMappingInput{
		{SourceFieldID: sourceB, TargetFieldID: targetB,
			ConverterKind: profilebackfillapp.MappingNormalize},
		{SourceFieldID: sourceA, TargetFieldID: targetA,
			ConverterKind: profilebackfillapp.MappingNormalize},
	}
	second := []ProfileMigrationMappingInput{first[1], first[0]}
	first, valid := canonicalProfileMigrationMappings(first)
	if !valid {
		t.Fatal("first mapping order rejected")
	}
	second, valid = canonicalProfileMigrationMappings(second)
	if !valid {
		t.Fatal("second mapping order rejected")
	}
	base := ProfileSchemaPublish{
		SchemaID: uuid.MustParse("14000000-0000-4000-8000-000000000001"),
		Resource: profileapp.ResourceUser, Version: 2,
		ExpectedActiveSchemaID: uuid.MustParse("14000000-0000-4000-8000-000000000002"),
		Fields:                 []ProfileFieldInput{},
	}
	base.MigrationMappings = first
	firstSHA, err := profilePublishRequestSHA(base)
	if err != nil {
		t.Fatal(err)
	}
	base.MigrationMappings = second
	secondSHA, err := profilePublishRequestSHA(base)
	if err != nil || firstSHA != secondSHA {
		t.Fatalf("canonical hashes differ: %x/%x/%v", firstSHA, secondSHA, err)
	}
	if _, valid := canonicalProfileMigrationMappings([]ProfileMigrationMappingInput{
		first[0], first[0],
	}); valid {
		t.Fatal("duplicate mapping passed")
	}
}

func TestProfileBackfillMutationViewExposesPendingPlan(t *testing.T) {
	planID := uuid.MustParse("15000000-0000-4000-8000-000000000001")
	view := profileBackfillMutationView(ProfileBackfillPublicationResult{
		SchemaID: uuid.MustParse("14000000-0000-4000-8000-000000000001"),
		PlanID:   planID, Revision: 1, AuthorizationClock: 2,
	})
	document, err := json.Marshal(view)
	if err != nil || view.BackfillPlanID == nil || *view.BackfillPlanID != planID ||
		view.PublicationState != "planned" {
		t.Fatalf("planned view = %s/%+v/%v", document, view, err)
	}
}
