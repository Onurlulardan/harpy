package environmentoperator

import (
	"errors"
	"strings"
	"testing"
)

type sqlStateStub struct {
	code string
}

func (stub sqlStateStub) Error() string {
	return "ERROR: tenant schema goose history is below the minimum schema version (SQLSTATE " + stub.code + ")"
}

func (stub sqlStateStub) SQLState() string {
	return stub.code
}

func TestResumableKeepsCauseStageAndSQLState(t *testing.T) {
	cause := sqlStateStub{code: "55000"}
	err := Resumable(
		"tenant_migration.refresh_onboarding", "onboarding_state_mismatch",
		"operation_id=11111111-1111-4111-8111-111111111111", cause,
	)
	if !errors.Is(err, ErrResumable) || !errors.Is(err, cause) {
		t.Fatalf("chain = %v", err)
	}
	diagnostic := Diagnose(err)
	if diagnostic.Stage != "tenant_migration.refresh_onboarding" ||
		diagnostic.ErrorCode != "onboarding_state_mismatch" ||
		diagnostic.SQLState != "55000" || ObservedStage(err) != diagnostic.Stage {
		t.Fatalf("diagnostic = %+v", diagnostic)
	}
	if !strings.Contains(err.Error(), "stage=tenant_migration.refresh_onboarding") ||
		!strings.Contains(err.Error(), "sqlstate=55000") {
		t.Fatalf("error = %s", err)
	}
}

func TestResumableRedactsCredentials(t *testing.T) {
	cause := errors.New("dial postgres://harpy_owner:supersecret@db.internal/harpy password=supersecret bearer abc.def")
	err := Resumable("deployment_policy.create", "deployment_policy_failed",
		"dsn=postgres://harpy:supersecret@db/harpy private key block", cause)
	diagnostic := Diagnose(err)
	combined := err.Error() + " " + diagnostic.Detail
	for _, secret := range []string{"supersecret", "abc.def", "harpy_owner:supersecret", "private key"} {
		if strings.Contains(combined, secret) {
			t.Fatalf("diagnostic leaked %q in %s", secret, combined)
		}
	}
	if !strings.Contains(err.Error(), "postgres://redacted") ||
		diagnostic.Detail != "redacted sensitive diagnostic" {
		t.Fatalf("redaction = %s detail=%s", err, diagnostic.Detail)
	}
}
