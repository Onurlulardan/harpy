package postgres

import (
	"bytes"
	"testing"
)

func TestCuratedSuccessorMigrationPreservesUTF8Translations(t *testing.T) {
	const migration = "20260802232100_create_layout_curated_definition_v2"
	for _, origin := range []string{"fresh", "v1_adopted"} {
		content := readLayoutMigration(t, origin, migration)
		for _, expected := range [][]byte{
			[]byte("'Oluştur'"),
			[]byte("'Güncelle'"),
			[]byte("'Ayrıntılar'"),
		} {
			if !bytes.Contains(content, expected) {
				t.Fatalf("%s migration is missing UTF-8 translation %q",
					origin, expected)
			}
		}
		for _, mojibake := range [][]byte{
			[]byte("OluÅŸtur"),
			[]byte("GÃ¼ncelle"),
			[]byte("AyrÄ±ntÄ±lar"),
		} {
			if bytes.Contains(content, mojibake) {
				t.Fatalf("%s migration contains mojibake %q", origin, mojibake)
			}
		}
	}
}
