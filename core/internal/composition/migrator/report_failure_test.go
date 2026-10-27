package migrator

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestReportFailureWritesOnlyStableClassification(t *testing.T) {
	var output bytes.Buffer
	reportFailure(&output, nil, "dependency_preflight", "verify")
	text := output.String()
	if !strings.Contains(text, `"error_class":"dependency_preflight"`) ||
		!strings.Contains(text, `"command":"verify"`) {
		t.Fatalf("stable failure fields missing: %s", text)
	}
	for _, forbidden := range []string{"postgres://", "SELECT ", `C:\\secret`, "/secret"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("failure output leaked %q: %s", forbidden, text)
		}
	}
}

func TestReportFailureCauseWritesPostgresConstraintWithoutSQL(t *testing.T) {
	var output bytes.Buffer
	cause := errors.New(
		`failed to execute SQL: SELECT platform.cutover_ensure_tenancy_authority_v1(id) ` +
			`FROM platform.domain_cutovers; ERROR: new row for relation "domain_cutovers" ` +
			`violates check constraint "domain_cutovers_domain_check" (SQLSTATE 23514)`,
	)
	reportFailureCause(&output, nil, "migration", "up", cause)
	text := output.String()
	if !strings.Contains(text, `"error_class":"migration"`) ||
		!strings.Contains(text, `domain_cutovers_domain_check`) ||
		!strings.Contains(text, `SQLSTATE 23514`) {
		t.Fatalf("postgres constraint detail missing: %s", text)
	}
	for _, forbidden := range []string{"postgres://", "SELECT ", `C:\\secret`, "/secret"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("failure output leaked %q: %s", forbidden, text)
		}
	}
}

func TestReportFailureStageWritesSafeStartupStage(t *testing.T) {
	var output bytes.Buffer
	reportFailureStage(
		&output, nil, "build_contract", tenantMigrationCommand, "store_bind",
	)
	text := output.String()
	for _, expected := range []string{
		`"error_class":"build_contract"`,
		`"command":"tenant-migration-executor"`,
		`"startup_stage":"store_bind"`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("stable startup field %s missing: %s", expected, text)
		}
	}
}
