package maintenance

import (
	"crypto/sha256"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func compileProfileMappings(
	target profileapp.Schema,
	mappings []profileMapping,
) ([]profileapp.MigrationMapping, error) {
	targets := make(map[uuid.UUID]struct{}, len(mappings))
	compiled := make([]profileapp.MigrationMapping, 0, len(mappings))
	for _, mapping := range mappings {
		targetField, found := target.Field(mapping.targetID)
		converter := profileapp.MigrationConverter(mapping.kind)
		if !found || targetField.State() != profileapp.FieldActive || !converter.Valid() {
			return nil, ErrRuntimeInvalid
		}
		semantic := targetField.SemanticSHA256()
		if len(mapping.semanticSHA) != sha256.Size ||
			!equalBytes(semantic[:], mapping.semanticSHA) {
			return nil, ErrRuntimeInvalid
		}
		targets[mapping.targetID] = struct{}{}
		compiled = append(compiled, profileapp.MigrationMapping{
			SourceFieldID: mapping.sourceID,
			TargetFieldID: mapping.targetID,
			Converter:     converter,
		})
	}
	for _, field := range target.Fields() {
		if field.State() == profileapp.FieldActive && field.Required() {
			if _, mapped := targets[field.ID()]; !mapped {
				return nil, ErrRuntimeInvalid
			}
		}
	}
	return compiled, nil
}

func convertProfileDocument(
	sourceSchema profileapp.Schema,
	targetSchema profileapp.Schema,
	mappings []profileapp.MigrationMapping,
	resource profileResource,
) (profileapp.Document, error) {
	return profileapp.MigrateDocument(
		sourceSchema, targetSchema, mappings, resource.document, resource.sha,
		maintenanceProfileNormalizer{},
	)
}

func equalBytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	var difference byte
	for index := range left {
		difference |= left[index] ^ right[index]
	}
	return difference == 0
}
