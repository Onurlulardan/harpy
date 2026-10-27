package iam

import (
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

func profileBackfillDocument(
	sourceSchema profileapp.Schema,
	targetSchema profileapp.Schema,
	mappings []profileapp.MigrationMapping,
	resource profilebackfillapp.SourceResource,
) (profileapp.Document, error) {
	return profileapp.MigrateDocument(
		sourceSchema, targetSchema, mappings, resource.Document,
		resource.SHA256[:], profileHTTPNormalizer{},
	)
}

func profileBackfillMigrationMappings(
	mappings []profilebackfillapp.Mapping,
) ([]profileapp.MigrationMapping, error) {
	shared := make([]profileapp.MigrationMapping, 0, len(mappings))
	for _, mapping := range mappings {
		converter := profileapp.MigrationConverter(mapping.ConverterKind)
		if !converter.Valid() {
			return nil, profilebackfillapp.ErrUnavailable
		}
		shared = append(shared, profileapp.MigrationMapping{
			SourceFieldID: mapping.SourceFieldID,
			TargetFieldID: mapping.TargetFieldID,
			Converter:     converter,
		})
	}
	return shared, nil
}
