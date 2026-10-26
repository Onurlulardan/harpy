package maintenance

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func (store *store) loadProfileSchema(
	ctx context.Context,
	claim maintenanceapp.Claim,
	planID uuid.UUID,
	id uuid.UUID,
) (profileapp.Schema, error) {
	values, err := claimValues(claim)
	if err != nil {
		return profileapp.Schema{}, err
	}
	rows, err := store.database.Query(
		ctx, readProfileSchemaSQL, append(values, planID, id)...,
	)
	if err != nil {
		return profileapp.Schema{}, fmt.Errorf("read profile schema: %w", err)
	}
	defer rows.Close()
	var header *profileSchemaHeader
	fields := make([]profileapp.Field, 0)
	for rows.Next() {
		current, field, fieldSHA, scanErr := scanProfileSchemaRow(rows)
		if scanErr != nil {
			return profileapp.Schema{}, scanErr
		}
		if header == nil {
			header = &current
		} else if !header.same(current) {
			return profileapp.Schema{}, ErrRuntimeInvalid
		}
		if field != nil {
			validationDigest := field.ValidationSHA256()
			if !bytes.Equal(validationDigest[:], fieldSHA) {
				return profileapp.Schema{}, ErrRuntimeInvalid
			}
			fields = append(fields, *field)
		}
	}
	if err := rows.Err(); err != nil {
		return profileapp.Schema{}, fmt.Errorf("iterate profile schema: %w", err)
	}
	if header == nil || len(header.definitionSHA) != sha256.Size ||
		(header.state != "active" && header.state != "retired" && header.state != "draft") {
		return profileapp.Schema{}, ErrRuntimeInvalid
	}
	state := profileapp.SchemaActive
	if header.state == "retired" {
		state = profileapp.SchemaRetired
	}
	return profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: header.id, Resource: header.resource, Version: header.version,
		State: state, Fields: fields,
		ExpectedDefinitionSHA256: header.definitionSHA,
	})
}

func scanProfileSchemaRow(rows interface{ Scan(...any) error }) (
	profileSchemaHeader, *profileapp.Field, []byte, error,
) {
	var header profileSchemaHeader
	var headerResource string
	var fieldID *uuid.UUID
	var key, kind, cardinality, classification, normalization, state *string
	var required *bool
	var validation *string
	var validationSHA []byte
	if err := rows.Scan(
		&header.id, &headerResource, &header.version, &header.state,
		&header.definitionSHA, &fieldID, &key, &kind, &cardinality,
		&required, &classification, &normalization, &state,
		&validation, &validationSHA,
	); err != nil {
		return header, nil, nil, fmt.Errorf("scan profile schema: %w", err)
	}
	header.resource = profileapp.ResourceKind(headerResource)
	if fieldID == nil {
		return header, nil, nil, nil
	}
	if key == nil || kind == nil || cardinality == nil || required == nil ||
		classification == nil || normalization == nil || state == nil || validation == nil {
		return header, nil, nil, ErrRuntimeInvalid
	}
	logicalType := profileapp.LogicalType(*kind)
	constraints, err := decodeProfileValidation([]byte(*validation), logicalType)
	if err != nil {
		return header, nil, nil, err
	}
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: *fieldID, Resource: header.resource, Key: *key,
		Type: logicalType, Cardinality: profileapp.Cardinality(*cardinality),
		Required: *required, Classification: profileapp.Classification(*classification),
		Normalization: profileapp.NormalizationProfile(*normalization),
		State:         profileapp.FieldState(*state), Validation: constraints,
	})
	if err != nil {
		return header, nil, nil, ErrRuntimeInvalid
	}
	return header, &field, validationSHA, nil
}
