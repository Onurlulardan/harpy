package runtimev1

import (
	"path/filepath"
	"testing"
)

func TestOpenAPIOperationsUseOneCategoryTag(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	for name, tag := range openAPIFileTags() {
		document := readOpenAPIDocument(t, filepath.Join(directory, name))
		assertOperationTags(t, name, document, tag)
	}
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	assertOperationTags(t, "core-api-v1.openapi.yaml", core, "Application")
	if _, found := core["x-tagGroups"]; found {
		t.Fatal("core API nests categories under another sidebar group")
	}
}

func assertOperationTags(t *testing.T, name string, value any, tag string) {
	t.Helper()
	switch current := value.(type) {
	case map[string]any:
		if current["operationId"] != nil {
			tags, _ := current["tags"].([]any)
			if len(tags) != 1 || tags[0] != tag {
				t.Fatalf("%s operation %v tags=%v want %s", name, current["operationId"], tags, tag)
			}
		}
		for _, child := range current {
			assertOperationTags(t, name, child, tag)
		}
	case []any:
		for _, child := range current {
			assertOperationTags(t, name, child, tag)
		}
	}
}

func openAPIFileTags() map[string]string {
	return map[string]string{
		"iam-auth-v1.paths.json": "Auth", "iam-lifecycle-public-v1.paths.json": "Auth",
		"iam-lifecycle-sessions-v1.paths.json": "Auth", "iam-lifecycle-account-v1.paths.json": "Auth",
		"iam-lifecycle-authenticators-v1.paths.json": "Auth", "iam-post-restore-v1.paths.json": "Auth",
		"iam-portal-users-v1.paths.json": "Auth", "iam-profile-v1.paths.json": "Profile",
		"iam-invitations-v1.paths.json": "Invitations", "iam-rbac-v1.paths.json": "Permissions",
		"iam-permission-requests-v1.paths.json": "Permissions", "iam-users-v1.paths.json": "Directory",
		"iam-roles-v1.paths.json": "Directory", "iam-organizations-v1.paths.json": "Directory",
		"billing-platform-campaigns-v2.paths.json": "Billing", "billing-platform-discounts-v2.paths.json": "Billing",
		"billing-platform-offerings-v2.paths.json": "Billing", "billing-platform-reports-v2.paths.json": "Billing",
		"billing-public-tenant-v2.paths.json": "Billing", "billing-webhook-v2.paths.json": "Billing",
		"documents-v2.paths.json": "Documents", "documents-v1-compat.paths.json": "Documents",
		"notifications-v1.paths.json": "Notifications", "platform-database-v1.paths.json": "Platform",
		"platform-operations-v1.paths.json": "Platform",
		"platform-recovery-v1.paths.json":   "Platform", "tenant-provisioning-v1.paths.json": "Tenancy",
		"tenant-settings-v1.paths.json": "Tenancy",
	}
}
