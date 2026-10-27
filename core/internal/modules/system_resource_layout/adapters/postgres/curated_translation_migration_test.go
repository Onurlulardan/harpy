package postgres

import (
	"bytes"
	"testing"
)

func TestCuratedSuccessorMigrationPreservesUTF8Translations(t *testing.T) {
	const migration = "20260802232100_create_layout_curated_definition_v2"
	content := readLayoutMigration(t, migration)
	for _, expected := range [][]byte{
		[]byte("'Oluştur'"),
		[]byte("'Güncelle'"),
		[]byte("'Ayrıntılar'"),
	} {
		if !bytes.Contains(content, expected) {
			t.Fatalf("fresh migration is missing UTF-8 translation %q", expected)
		}
	}
	for _, mojibake := range [][]byte{
		[]byte("OluÅŸtur"),
		[]byte("GÃ¼ncelle"),
		[]byte("AyrÄ±ntÄ±lar"),
	} {
		if bytes.Contains(content, mojibake) {
			t.Fatalf("fresh migration contains mojibake %q", mojibake)
		}
	}
}
