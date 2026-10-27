package iam

import "testing"

func TestDocumentsAccessPermissionsMapOnlyExactOperations(t *testing.T) {
	for key, target := range documentsAccessTargets {
		if target.resource == "" || target.action == "" {
			t.Fatalf("Documents access target %q is incomplete", key)
		}
	}
	if len(documentsAccessTargets) != 32 {
		t.Fatalf("Documents access target count = %d", len(documentsAccessTargets))
	}
	if _, ok := resolveDocumentsAccess(
		"documents.folder.read", "documents.folder.delete"); ok {
		t.Fatal("Documents read permission authorized a delete operation")
	}
	if !documentsOperationRequiresStepUp("documents.item.purge") ||
		documentsOperationRequiresStepUp("documents.item.cancel_purge") {
		t.Fatal("Documents step-up operation set differs from the contract")
	}
}
