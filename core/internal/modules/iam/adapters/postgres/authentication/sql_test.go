package postgres

import (
	"strings"
	"testing"
)

func TestAuthenticationTenantSQLUsesPGFirstRuntimeTenantSlot(t *testing.T) {
	statements := map[string]string{
		"login":                       loginCredentialSQL,
		"password":                    usePasswordSQL,
		"session":                     issueSessionSQL,
		"refresh":                     refreshCredentialSQL,
		"rotation":                    rotateRefreshSQL,
		"reauthentication challenge":  createReauthenticationChallengeSQL,
		"reauthentication credential": getReauthenticationCredentialSQL,
		"reauthentication rejection":  rejectReauthenticationSQL,
		"reauthentication completion": completeReauthenticationSQL,
		"session revocation":          revokeSessionSQL,
		"recovery credential":         resolveRecoveryCredentialSQL,
		"recovery consumption":        consumeRecoverySQL,
		"initial credential":          initialPasswordCredentialSQL,
		"initial start":               initialPasswordStartSQL,
		"initial continuation":        initialPasswordContinuationSQL,
		"initial rejection":           initialPasswordRejectSQL,
		"initial completion":          initialPasswordCompleteSQL,
	}
	for name, statement := range statements {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(statement,
				"(pg_catalog.current_schemas(false))[2]::name") ||
				strings.Contains(statement, "current_schema()") {
				t.Fatal("tenant SQL does not resolve the PG-first tenant slot")
			}
			for _, forbidden := range []string{
				"INSERT ", "UPDATE ", "DELETE ", "platform.", "pg_temp",
			} {
				if strings.Contains(strings.ToUpper(statement), strings.ToUpper(forbidden)) {
					t.Fatalf("tenant SQL contains forbidden %q", forbidden)
				}
			}
		})
	}
}

func TestAuthenticationPlatformSQLUsesOnlyNarrowLocatorRoutines(t *testing.T) {
	tests := map[string]struct {
		statement string
		routine   string
	}{
		"create":  {createRefreshLocatorSQL, "platform.iam_refresh_locator_create_v1("},
		"advance": {advanceRefreshLocatorSQL, "platform.iam_refresh_locator_advance_v1("},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if strings.Count(test.statement, "platform.") != 1 ||
				!strings.Contains(test.statement, test.routine) {
				t.Fatal("platform SQL does not use exactly one narrow locator routine")
			}
		})
	}
}

func TestAuthenticationUsesStateBoundRefreshRoutines(t *testing.T) {
	if !strings.Contains(refreshCredentialSQL, "iam_refresh_credential_v3(") ||
		strings.Contains(refreshCredentialSQL, "iam_refresh_credential_v2(") ||
		!strings.Contains(rotateRefreshSQL, "iam_refresh_rotate_v4(") ||
		strings.Contains(rotateRefreshSQL, "iam_refresh_rotate_v3(") {
		t.Fatal("authentication SQL uses a stale refresh locator binding")
	}
}
