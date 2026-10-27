package iam

import (
	"strings"
	"testing"
)

func TestInitialAdminSetupReissueMigrationRevokesThenReplacesAtomically(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	var routine string
	for _, document := range documents {
		if strings.Contains(document.Path(), "/tenant/fresh/") &&
			strings.Contains(document.Path(),
				"20260722140104_create_iam_initial_admin_setup_reissue_v1") {
			routine = string(document.Content())
			break
		}
	}
	required := []string{
		"v_setup.state<>'pending'", "app_user.identity_kind='human'",
		"platform.tenancy_customer_setup_reissue_v1(",
		"SET state='revoked'", "DELETE FROM iam_lifecycle_delivery_artifacts",
		"INSERT INTO iam_one_time_tokens", "INSERT INTO iam_lifecycle_delivery_artifacts",
		"generation=generation+1", "platform_async_initial_admin_setup_reissue_enqueue_v1(",
	}
	for _, token := range required {
		if !strings.Contains(routine, token) {
			t.Fatalf("initial admin setup reissue routine lacks %q", token)
		}
	}
	revoke := strings.Index(routine, "SET state='revoked'")
	deleteArtifact := strings.Index(routine, "DELETE FROM iam_lifecycle_delivery_artifacts")
	insertToken := strings.Index(routine, "INSERT INTO iam_one_time_tokens")
	insertArtifact := strings.Index(routine, "INSERT INTO iam_lifecycle_delivery_artifacts")
	if revoke < 0 || !(revoke < deleteArtifact && deleteArtifact < insertToken &&
		insertToken < insertArtifact) || strings.Contains(routine, "COMMIT") {
		t.Fatal("initial admin setup reissue replacement order is not atomic")
	}
}
