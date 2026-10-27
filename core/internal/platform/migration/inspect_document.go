package migration

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

type DocumentMetadata = postgresv1.MigrationSourceMetadata

func InspectDocument(document Document) (DocumentMetadata, error) {
	return postgresv1.InspectMigrationSource(document.content)
}
