package migrator

import "testing"

func TestDeployMigrationDocumentsDoNotLoadCopiedSchema(t *testing.T) {
	documents, err := deployMigrationDocuments()
	if err == nil || documents != nil {
		t.Fatalf("copied schema migrations leaked: documents=%d err=%v", len(documents), err)
	}
}
