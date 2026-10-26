package postgres

import (
	"strings"
	"testing"
)

func TestInvitationSQLUsesPGFirstTenantAndClosedRoutines(t *testing.T) {
	documents := map[string]string{
		"create": invitationCreateSQL, "read": invitationReadSQL,
		"list": invitationListSQL, "revoke": invitationRevokeSQL,
		"expire": invitationExpireSQL, "accept": invitationAcceptSQL,
	}
	for name, document := range documents {
		if !strings.Contains(document, "(pg_catalog.current_schemas(false))[2]::name") {
			t.Fatalf("%s SQL does not resolve the PG-first tenant schema", name)
		}
		if strings.Contains(document, "pg_catalog.current_schema()") ||
			strings.Contains(document, "iam_invitations AS") {
			t.Fatalf("%s SQL escaped the closed invitation routine boundary", name)
		}
	}
}
