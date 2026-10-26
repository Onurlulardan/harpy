package iam

import (
	"sort"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func canonicalProfileMigrationMappings(
	input []ProfileMigrationMappingInput,
) ([]ProfileMigrationMappingInput, bool) {
	if len(input) > 500 {
		return nil, false
	}
	if len(input) == 0 {
		return nil, true
	}
	result := append([]ProfileMigrationMappingInput(nil), input...)
	for _, mapping := range result {
		if !profileMappingID(mapping.SourceFieldID) ||
			!profileMappingID(mapping.TargetFieldID) ||
			mapping.ConverterKind != profilebackfillapp.MappingCopy &&
				mapping.ConverterKind != profilebackfillapp.MappingNormalize ||
			mapping.ConverterKind == profilebackfillapp.MappingCopy &&
				mapping.SourceFieldID != mapping.TargetFieldID ||
			mapping.ConverterKind == profilebackfillapp.MappingNormalize &&
				mapping.SourceFieldID == mapping.TargetFieldID {
			return nil, false
		}
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].TargetFieldID != result[right].TargetFieldID {
			return result[left].TargetFieldID.String() < result[right].TargetFieldID.String()
		}
		return result[left].SourceFieldID.String() < result[right].SourceFieldID.String()
	})
	sources, targets := map[uuid.UUID]bool{}, map[uuid.UUID]bool{}
	for _, mapping := range result {
		if sources[mapping.SourceFieldID] || targets[mapping.TargetFieldID] {
			return nil, false
		}
		sources[mapping.SourceFieldID], targets[mapping.TargetFieldID] = true, true
	}
	return result, true
}

func profileMappingID(id uuid.UUID) bool {
	return id != uuid.Nil && id.Variant() == uuid.RFC4122
}
