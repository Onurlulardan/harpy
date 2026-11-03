package operator

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func TestInitializeFailureRecordKeepsStageAndSQLState(t *testing.T) {
	var stderr bytes.Buffer
	err := environmentoperator.Resumable(
		"tenant_migration.refresh_onboarding", "onboarding_state_mismatch",
		"operation_id=11111111-1111-4111-8111-111111111111 onboarding_state=migrating",
		sqlStateStub{code: "55000"},
	)
	writeInitializeFailure(&stderr, "initialize", processcode.OperatorResumable, err)
	record := stderr.String()
	for _, part := range []string{
		`"stage":"tenant_migration.refresh_onboarding"`,
		`"error_code":"onboarding_state_mismatch"`,
		`"sqlstate":"55000"`,
		`"exit_code":5`,
		"onboarding_state=migrating",
	} {
		if !strings.Contains(record, part) {
			t.Fatalf("record %q missing %s", record, part)
		}
	}
}

type sqlStateStub struct {
	code string
}

func (stub sqlStateStub) Error() string { return "sql failed" }

func (stub sqlStateStub) SQLState() string { return stub.code }

func TestInitializeFailureRecordOmitsEmptyDiagnostics(t *testing.T) {
	var stderr bytes.Buffer
	writeInitializeFailure(&stderr, "configuration", processcode.InvalidStartup,
		errors.New("configuration is invalid"))
	record := stderr.String()
	if strings.Contains(record, "error_code") || strings.Contains(record, "sqlstate") ||
		!strings.Contains(record, `"stage":"configuration"`) {
		t.Fatalf("record = %s", record)
	}
}
