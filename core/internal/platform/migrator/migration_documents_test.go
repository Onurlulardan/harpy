package migrator

import "testing"

func TestMigrationDocumentsStayUnembedded(t *testing.T) {
	documents, err := MigrationDocuments()
	if err == nil || documents != nil {
		t.Fatalf("copied schema migrations leaked: documents=%v err=%v", documents, err)
	}
}
