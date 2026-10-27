//go:build integration

package integrationtest

import (
	"testing"

	identifierapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

const (
	recoveryUserTwoID       = "51000000-0000-4000-8000-000000000002"
	recoveryUserThreeID     = "51000000-0000-4000-8000-000000000003"
	recoveryUserV4ID        = "41000000-0000-4000-8000-000000000004"
	recoveryReplacementV4ID = "43000000-0000-4000-8000-000000000004"
	recoveryOrganizationV2  = "51000000-0000-4000-8000-000000000100"
)

type recoveryTextNormalizer struct {
	drift bool
	calls map[string]int
}

func (normalizer recoveryTextNormalizer) NormalizeEmailV1(value string) (string, error) {
	result, err := identifierapp.NormalizeEmailV1(value)
	if err != nil {
		return "", err
	}
	if normalizer.drift && result.Normalized() == "e@f.co" {
		normalizer.calls[result.Normalized()]++
		if normalizer.calls[result.Normalized()]%2 == 0 {
			return "drift@example.com", nil
		}
	}
	return result.Normalized(), nil
}

func (recoveryTextNormalizer) NormalizePortalV1(value string) (string, error) {
	result, err := identifierapp.NormalizePortalV1(value)
	if err != nil {
		return "", err
	}
	return result.Normalized(), nil
}

func recoveryUserTargetV3(t *testing.T) profileapp.Schema {
	t.Helper()
	retired, err := profileapp.RetiredField(
		publicationCurrentField(t, publicationSourceField, "display_name"),
	)
	if err != nil {
		t.Fatal(err)
	}
	fields := []profileapp.Field{
		retired,
		publicationCurrentField(t, publicationOptionalID, "nickname"),
		recoveryEmailField(t, publicationReplacement, "contact_alias_v2", true),
	}
	return recoverySchema(t, publicationTargetV3ID, profileapp.ResourceUser, 3, fields)
}

func recoveryUserTargetV4(t *testing.T) profileapp.Schema {
	t.Helper()
	v3 := recoveryUserTargetV3(t)
	retiredDisplay, _ := v3.Field(uuid.MustParse(publicationSourceField))
	retiredEmail, _ := v3.Field(uuid.MustParse(publicationReplacement))
	retiredEmail, err := profileapp.RetiredField(retiredEmail)
	if err != nil {
		t.Fatal(err)
	}
	fields := []profileapp.Field{
		retiredDisplay,
		publicationCurrentField(t, publicationOptionalID, "nickname"),
		retiredEmail,
		recoveryEmailField(t, recoveryReplacementV4ID, "contact_alias_v3", true),
	}
	return recoverySchema(t, recoveryUserV4ID, profileapp.ResourceUser, 4, fields)
}

func recoveryOrganizationTargetV3(t *testing.T) profileapp.Schema {
	t.Helper()
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: uuid.MustParse(publicationEmptyField), Resource: profileapp.ResourceOrganization,
		Key: "nickname", Type: profileapp.LogicalString,
		Cardinality:    profileapp.CardinalitySingle,
		Classification: profileapp.ClassificationInternal,
		Normalization:  profileapp.NormalizationNoneV1, State: profileapp.FieldActive,
		Validation: profileapp.ValidationInput{MaximumBytes: 16},
	})
	if err != nil {
		t.Fatal(err)
	}
	return recoverySchema(
		t, publicationEmptyV3, profileapp.ResourceOrganization, 3,
		[]profileapp.Field{field},
	)
}

func recoveryEmailField(t *testing.T, id, key string, required bool) profileapp.Field {
	t.Helper()
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: uuid.MustParse(id), Resource: profileapp.ResourceUser, Key: key,
		Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
		Required: required, Classification: profileapp.ClassificationInternal,
		Normalization: profileapp.NormalizationEmailV1, State: profileapp.FieldActive,
		Validation: profileapp.ValidationInput{MaximumBytes: 128},
	})
	if err != nil {
		t.Fatal(err)
	}
	return field
}

func recoverySchema(
	t *testing.T, id string, resource profileapp.ResourceKind,
	version int64, fields []profileapp.Field,
) profileapp.Schema {
	t.Helper()
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(id), Resource: resource, Version: version,
		State: profileapp.SchemaActive, Fields: fields,
	})
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func recoveryMappingsV3() []profilebackfillapp.Mapping {
	return []profilebackfillapp.Mapping{
		{SourceFieldID: uuid.MustParse(publicationSourceField),
			TargetFieldID: uuid.MustParse(publicationReplacement),
			ConverterKind: profilebackfillapp.MappingNormalize},
		{SourceFieldID: uuid.MustParse(publicationOptionalID),
			TargetFieldID: uuid.MustParse(publicationOptionalID),
			ConverterKind: profilebackfillapp.MappingCopy},
	}
}

func recoveryMappingsV4() []profilebackfillapp.Mapping {
	return []profilebackfillapp.Mapping{
		{SourceFieldID: uuid.MustParse(publicationReplacement),
			TargetFieldID: uuid.MustParse(recoveryReplacementV4ID),
			ConverterKind: profilebackfillapp.MappingNormalize},
		{SourceFieldID: uuid.MustParse(publicationOptionalID),
			TargetFieldID: uuid.MustParse(publicationOptionalID),
			ConverterKind: profilebackfillapp.MappingCopy},
	}
}

func recoveryOrganizationMappings() []profilebackfillapp.Mapping {
	return []profilebackfillapp.Mapping{{
		SourceFieldID: uuid.MustParse(publicationEmptyField),
		TargetFieldID: uuid.MustParse(publicationEmptyField),
		ConverterKind: profilebackfillapp.MappingCopy,
	}}
}
