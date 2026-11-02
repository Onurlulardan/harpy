package postgres

import (
	"os"
	"path/filepath"
	"testing"
)

var securityEpochMigrationNames = map[string][]string{
	"platform": {
		"20260722022000_create_iam_security_epochs",
		"20260722022001_guard_iam_security_epochs",
		"20260722022002_create_iam_security_epoch_initialize_v1",
		"20260725123500_grant_iam_installation_security_epoch_runtime",
	},
	"tenant": {
		"20260722022100_create_iam_security_epoch_current_v1",
		"20260722022101_bind_iam_credentials_security_epoch",
		"20260722022102_guard_iam_credential_security_epoch",
		"20260722022103_create_iam_session_issue_v2",
		"20260722022104_create_iam_refresh_credential_v3",
		"20260722022105_create_iam_refresh_rotate_v4",
		"20260722022106_create_iam_current_session_read_v2",
		"20260722022107_grant_iam_security_epoch_runtime_v1",
		"20260722022108_create_iam_security_epoch_assert_v1",
	},
}

func readSecurityEpochMigration(
	t *testing.T, scope, origin, name string,
) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "migrations", scope, origin,
		name+"."+origin+".sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}
