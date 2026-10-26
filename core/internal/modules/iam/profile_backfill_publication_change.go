package iam

import (
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func profileBackfillRequired(
	source profileapp.Schema,
	target profileapp.Schema,
	input []ProfileMigrationMappingInput,
) (bool, error) {
	declared, err := profileDeclaredMappings(source, target, input)
	if err != nil {
		return false, err
	}
	mappedSources := make(map[uuid.UUID]profilebackfillapp.Mapping, len(declared))
	mappedTargets := make(map[uuid.UUID]bool, len(declared))
	for _, mapping := range declared {
		mappedSources[mapping.SourceFieldID] = mapping
		mappedTargets[mapping.TargetFieldID] = true
	}
	required := len(declared) > 0
	for _, sourceField := range source.Fields() {
		targetField, found := target.Field(sourceField.ID())
		if !found {
			return false, ErrProfileBackfillUnmappable
		}
		if sourceField.SemanticSHA256() != targetField.SemanticSHA256() {
			return false, ErrProfileBackfillUnmappable
		}
		if sourceField.State() == profileapp.FieldRetired {
			if targetField.State() != profileapp.FieldRetired ||
				sourceField.ValidationSHA256() != targetField.ValidationSHA256() {
				return false, ErrProfileBackfillUnmappable
			}
			continue
		}
		if targetField.State() == profileapp.FieldRetired {
			mapping, replaced := mappedSources[sourceField.ID()]
			if !replaced || mapping.TargetFieldID == sourceField.ID() {
				return false, ErrProfileBackfillUnmappable
			}
			required = true
			continue
		}
		if sourceField.ValidationSHA256() != targetField.ValidationSHA256() {
			required = true
		}
	}
	for _, targetField := range target.Fields() {
		if _, found := source.Field(targetField.ID()); found {
			continue
		}
		if targetField.State() != profileapp.FieldActive ||
			targetField.Required() && !mappedTargets[targetField.ID()] {
			return false, ErrProfileBackfillUnmappable
		}
	}
	return required, nil
}

func profileDeclaredMappings(
	source profileapp.Schema,
	target profileapp.Schema,
	input []ProfileMigrationMappingInput,
) ([]profilebackfillapp.Mapping, error) {
	canonical, valid := canonicalProfileMigrationMappings(input)
	if !valid {
		return nil, ErrProfileBackfillUnmappable
	}
	result := make([]profilebackfillapp.Mapping, 0, len(canonical))
	for _, candidate := range canonical {
		sourceField, sourceFound := source.Field(candidate.SourceFieldID)
		targetField, targetFound := target.Field(candidate.TargetFieldID)
		if !sourceFound || !targetFound ||
			sourceField.State() != profileapp.FieldActive ||
			targetField.State() != profileapp.FieldActive {
			return nil, ErrProfileBackfillUnmappable
		}
		switch candidate.ConverterKind {
		case profilebackfillapp.MappingCopy:
			if sourceField.ID() != targetField.ID() ||
				sourceField.SemanticSHA256() != targetField.SemanticSHA256() {
				return nil, ErrProfileBackfillUnmappable
			}
		case profilebackfillapp.MappingNormalize:
			if sourceField.ID() == targetField.ID() ||
				sourceField.Type() != targetField.Type() ||
				sourceField.Cardinality() != targetField.Cardinality() {
				return nil, ErrProfileBackfillUnmappable
			}
			retiredSource, retained := target.Field(sourceField.ID())
			if !retained || retiredSource.State() != profileapp.FieldRetired ||
				retiredSource.SemanticSHA256() != sourceField.SemanticSHA256() {
				return nil, ErrProfileBackfillUnmappable
			}
		default:
			return nil, ErrProfileBackfillUnmappable
		}
		result = append(result, profilebackfillapp.Mapping{
			SourceFieldID: sourceField.ID(), TargetFieldID: targetField.ID(),
			ConverterKind: candidate.ConverterKind,
		})
	}
	return result, nil
}
