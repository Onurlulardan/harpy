package passwordenrollmentpostgres

import (
	"strings"
	"testing"
)

func TestCreateDelegatesOnlyAtomicRoutine(t *testing.T) {
	if !strings.Contains(enrollSQL, "FROM iam_identity_password_enroll_v1(") ||
		!strings.Contains(enrollSQL, "(pg_catalog.current_schemas(false))[2]::name") ||
		strings.Contains(strings.ToUpper(enrollSQL), "INSERT ") {
		t.Fatal("password enrollment SQL bypasses the atomic routine")
	}
}
