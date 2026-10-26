package profilepostgres

import (
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func migrateProfileWrite(
	contract profileWriteContract,
	source profileapp.Document,
) (profileapp.Document, error) {
	wire := source.CanonicalJSON()
	defer clear(wire)
	digest := source.SHA256()
	return profileapp.MigrateDocument(
		contract.source, contract.target, contract.mappings,
		wire, digest[:], profileNormalizer{},
	)
}

func migrationTargetForWrite(
	contract profileWriteContract,
	current profileWriteResource,
) (profileapp.Document, error) {
	if current.actualSchemaID == contract.target.ID() {
		return current.document, nil
	}
	mappings, err := profileapp.SelectMigrationMappingsForSource(
		current.actualSchema, contract.target, contract.mappings,
	)
	if err != nil {
		return profileapp.Document{}, err
	}
	wire := current.document.CanonicalJSON()
	defer clear(wire)
	digest := current.document.SHA256()
	return profileapp.MigrateDocument(
		current.actualSchema, contract.target, mappings,
		wire, digest[:], profileNormalizer{},
	)
}
