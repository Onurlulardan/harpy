package iam

import (
	"strconv"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func profilePublishedSchema(input ProfileSchemaPublish) (profileapp.Schema, error) {
	if input.SchemaID.Version() == 0 || !input.Resource.Valid() || input.Version < 2 ||
		input.ExpectedActiveSchemaID.Version() == 0 || input.Fields == nil {
		return profileapp.Schema{}, ErrProfileHTTPInvalid
	}
	fields := make([]profileapp.Field, len(input.Fields))
	for index, candidate := range input.Fields {
		field, err := profileField(input.Resource, candidate)
		if err != nil {
			return profileapp.Schema{}, err
		}
		fields[index] = field
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: input.SchemaID, Resource: input.Resource, Version: input.Version,
		State: profileapp.SchemaActive, Fields: fields,
	})
	if err != nil {
		return profileapp.Schema{}, profileApplicationError(err)
	}
	return schema, nil
}

func profileField(
	resource profileapp.ResourceKind,
	input ProfileFieldInput,
) (profileapp.Field, error) {
	minimum, err := profileOptionalInt64(input.Validation.IntegerMinimum)
	if err != nil {
		return profileapp.Field{}, err
	}
	maximum, err := profileOptionalInt64(input.Validation.IntegerMaximum)
	if err != nil {
		return profileapp.Field{}, err
	}
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: input.ID, Resource: resource, Key: input.Key, Type: input.Type,
		Cardinality: input.Cardinality, Required: input.Required,
		Classification: input.Classification, Normalization: input.Normalization,
		State: input.State, Validation: profileapp.ValidationInput{
			MaximumBytes:   input.Validation.MaximumBytes,
			MaximumItems:   input.Validation.MaximumItems,
			IntegerMinimum: minimum, IntegerMaximum: maximum,
			DecimalPrecision: input.Validation.DecimalPrecision,
			DecimalScale:     input.Validation.DecimalScale,
			EnumValues:       input.Validation.EnumValues,
		},
	})
	if err != nil {
		return profileapp.Field{}, profileApplicationError(err)
	}
	return field, nil
}

func profileOptionalInt64(value *string) (*int64, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(*value, 10, 64)
	if err != nil || strconv.FormatInt(parsed, 10) != *value {
		return nil, ErrProfileHTTPInvalid
	}
	return &parsed, nil
}
