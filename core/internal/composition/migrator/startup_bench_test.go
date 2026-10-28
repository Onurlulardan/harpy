package migrator

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func BenchmarkAPISchemaContractSQL(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		documents, err := FreshMigrationDocuments()
		if err != nil {
			b.Fatal(err)
		}
		if _, _, err = schemacontract.LoadStartup(documents); err != nil {
			b.Fatal(err)
		}
	}
}
