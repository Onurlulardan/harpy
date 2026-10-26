//go:build integration

package integrationtest

import (
	"testing"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func publicationClaims(t *testing.T, tenant publicationTenant) accesstoken.Claims {
	t.Helper()
	profile, err := accesstoken.NewProfile(
		"http://localhost:8000", "http://localhost:5173", true,
	)
	if err != nil {
		t.Fatal(err)
	}
	issued := time.Date(2026, 7, 16, 11, 0, 0, 0, time.UTC)
	claims, err := accesstoken.NewClaims(profile, accesstoken.ClaimsInput{
		Subject: uuid.MustParse(publicationActorID),
		Session: uuid.MustParse("52000000-0000-4000-8000-000000000001"),
		Tenant:  uuid.MustParse(tenant.id),
		Client:  accesstoken.ClientOperatorConsole, IssuedAt: issued,
		JTI:          uuid.MustParse("53000000-0000-4000-8000-000000000001"),
		TokenVersion: 1, AuthorizationVersion: 1,
		PlatformAuthorizationClock: 7, SecurityEpoch: [32]byte{1},
		AccessMode:         accesstoken.AccessModeOperator,
		AuthenticationTime: issued, Assurance: accesstoken.AssuranceBaseline,
		Methods: []accesstoken.AuthMethod{accesstoken.AuthMethodPassword},
	})
	if err != nil {
		t.Fatal(err)
	}
	return claims
}

func publicationOptionalInput() iammodule.ProfileSchemaPublish {
	return iammodule.ProfileSchemaPublish{
		IdempotencyKey: "profile-optional-v2",
		SchemaID:       uuid.MustParse(publicationOptionalV2ID),
		Resource:       profileapp.ResourceUser, Version: 2,
		ExpectedActiveSchemaID: uuid.MustParse(publicationSourceID),
		Fields: []iammodule.ProfileFieldInput{
			publicationField(publicationSourceField, "display_name", 8),
			publicationField(publicationOptionalID, "nickname", 8),
		},
	}
}

func publicationBackfillInput() iammodule.ProfileSchemaPublish {
	retired := publicationField(publicationSourceField, "display_name", 64)
	retired.State = profileapp.FieldRetired
	replacement := publicationField(publicationReplacement, "contact_alias_v2", 128)
	replacement.Required = true
	replacement.Normalization = profileapp.NormalizationEmailV1
	return iammodule.ProfileSchemaPublish{
		IdempotencyKey: "profile-replacement-v3",
		SchemaID:       uuid.MustParse(publicationTargetV3ID),
		Resource:       profileapp.ResourceUser, Version: 3,
		ExpectedActiveSchemaID: uuid.MustParse(publicationOptionalV2ID),
		Fields: []iammodule.ProfileFieldInput{
			retired,
			publicationField(publicationOptionalID, "nickname", 8),
			replacement,
		},
		MigrationMappings: []iammodule.ProfileMigrationMappingInput{
			{
				SourceFieldID: uuid.MustParse(publicationSourceField),
				TargetFieldID: uuid.MustParse(publicationReplacement),
				ConverterKind: "copy_normalize_v1",
			},
		},
	}
}

func publicationInvalidBackfillInput() iammodule.ProfileSchemaPublish {
	input := publicationBackfillInput()
	input.IdempotencyKey = "profile-invalid-row-v3"
	for index := range input.Fields {
		if input.Fields[index].ID == uuid.MustParse(publicationReplacement) {
			input.Fields[index].Validation.MaximumBytes = 3
		}
	}
	return input
}

func publicationField(id, key string, maximumBytes int) iammodule.ProfileFieldInput {
	classification := profileapp.ClassificationInternal
	if key == "display_name" {
		classification = profileapp.ClassificationPublic
	}
	return iammodule.ProfileFieldInput{
		ID: uuid.MustParse(id), Key: key, Type: profileapp.LogicalString,
		Cardinality: profileapp.CardinalitySingle, Required: false,
		Classification: classification,
		Normalization:  profileapp.NormalizationNoneV1,
		State:          profileapp.FieldActive,
		Validation:     iammodule.ProfileValidationInput{MaximumBytes: maximumBytes},
	}
}
