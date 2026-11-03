package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestAdminConnectErrorNamesStaleSecretGeneration(t *testing.T) {
	wrapped := adminConnectError(errors.New(
		`failed SASL auth: FATAL: password authentication failed for user "local_bootstrap_admin" (SQLSTATE 28P01)`,
	))
	typed := adminConnectError(&pgconn.PgError{Code: "28P01", Message: "password authentication failed"})
	for _, err := range []error{wrapped, typed} {
		text := err.Error()
		if !strings.Contains(text, "previous secret generation") ||
			!strings.Contains(text, "task local:destroy") {
			t.Fatalf("stale password error = %q", text)
		}
	}
	other := adminConnectError(errors.New("dial tcp: connection refused"))
	if strings.Contains(other.Error(), "task local:destroy") ||
		!strings.Contains(other.Error(), "connect admin DSN:") {
		t.Fatalf("unrelated connect error = %q", other)
	}
}
