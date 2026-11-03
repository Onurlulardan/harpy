package migrator

import "github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"

func loadSchemaContract() (schemacontract.EmbeddedContract, error) {
	return schemacontract.LoadBoundStartup()
}
