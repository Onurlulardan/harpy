package aiplatformowner

import (
	"context"
	"strings"
	"testing"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestAIPlatformOwnerSQLUsesExactSessionAndOwnerLockRoutines(t *testing.T) {
	for _, marker := range []string{
		"iam_current_session_read_v6", "$1::uuid", "$2::uuid",
		"result_token_version = $3::bigint",
		"result_authorization_version = $4::bigint",
	} {
		if !strings.Contains(currentSessionSQL, marker) {
			t.Fatalf("current session SQL omits %q", marker)
		}
	}
	for _, marker := range []string{
		"platform.iam_platform_owner_context_lock_v1",
		"$1::uuid", "$2::uuid", "$3::bigint",
	} {
		if !strings.Contains(currentOwnerSQL, marker) {
			t.Fatalf("current owner SQL omits %q", marker)
		}
	}
	for _, forbidden := range []string{
		"platform.ai.analytics.read", "iam_platform_authorize_v1", "capability",
	} {
		if strings.Contains(currentSessionSQL+currentOwnerSQL, forbidden) {
			t.Fatalf("AI current-owner SQL contains generic authority %q", forbidden)
		}
	}
}

func TestAIPlatformOwnerReaderSealsIAMCommandScope(t *testing.T) {
	port, err := New()
	if err != nil {
		t.Fatal(err)
	}
	bound, ok := port.(*reader)
	if !ok || bound.owner != "iam" || bound.mode != platformtx.Command {
		t.Fatal("AI current-owner reader is not sealed to IAM command scope")
	}
	if _, err := port.Read(context.Background(), platformtx.TenantScope{}, Input{}); err != ErrRejected {
		t.Fatalf("unowned scope error=%v", err)
	}
}

func TestAIPlatformOwnerReaderBindsExactInputs(t *testing.T) {
	input := Input{
		TenantID:             uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		ActorID:              uuid.MustParse("20000000-0000-4000-8000-000000000001"),
		SessionID:            uuid.MustParse("30000000-0000-4000-8000-000000000001"),
		ExpectedTokenVersion: 3, ExpectedAuthorizationVersion: 5,
		ExpectedPlatformClock: 7,
	}
	fixture := &ownerQueryFixture{input: input}
	bound := &boundReader{database: fixture}
	evidence, err := bound.readSession(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if err = bound.readOwner(context.Background(), input, &evidence); err != nil {
		t.Fatal(err)
	}
	if fixture.calls != 2 || !evidence.validShape() ||
		evidence.PlatformSubjectID != evidence.OwnerSubjectID {
		t.Fatalf("evidence=%+v calls=%d", evidence, fixture.calls)
	}
}

type ownerQueryFixture struct {
	input Input
	calls int
}

func (fixture *ownerQueryFixture) QueryRow(
	_ context.Context, query string, arguments ...any,
) platformpostgres.Row {
	fixture.calls++
	if fixture.calls == 1 {
		return ownerRow(func(destinations ...any) error {
			if query != currentSessionSQL || len(arguments) != 4 ||
				arguments[0] != fixture.input.ActorID || arguments[1] != fixture.input.SessionID ||
				arguments[2] != int64(3) || arguments[3] != int64(5) {
				return ErrUnavailable
			}
			setSessionEvidence(destinations, fixture.input)
			return nil
		})
	}
	return ownerRow(func(destinations ...any) error {
		if query != currentOwnerSQL || len(arguments) != 3 ||
			arguments[0] != fixture.input.TenantID || arguments[1] != fixture.input.ActorID ||
			arguments[2] != int64(7) {
			return ErrUnavailable
		}
		setOwnerEvidence(destinations)
		return nil
	})
}

type ownerRow func(...any) error

func (row ownerRow) Scan(destinations ...any) error { return row(destinations...) }

func setSessionEvidence(destinations []any, input Input) {
	*destinations[0].(*uuid.UUID), *destinations[1].(*uuid.UUID) = input.ActorID, input.SessionID
	*destinations[2].(*string), *destinations[3].(*string) =
		"harpy-operator-console-v1", "operator_console"
	*destinations[4].(*uuid.NullUUID) = uuid.NullUUID{}
	*destinations[5].(*int64), *destinations[6].(*int64) = 3, 5
	*destinations[7].(*string), *destinations[8].(*string) = "onpremise", "onpremise_primary"
	subject := uuid.MustParse("40000000-0000-4000-8000-000000000001")
	*destinations[9].(*uuid.NullUUID) = uuid.NullUUID{UUID: subject, Valid: true}
	*destinations[10].(*int64) = 7
	*destinations[11].(*[]byte) = append([]byte(nil), append([]byte{1}, make([]byte, 31)...)...)
}

func setOwnerEvidence(destinations []any) {
	*destinations[0].(*uuid.UUID) = uuid.MustParse("50000000-0000-4000-8000-000000000001")
	*destinations[1].(*string) = "onpremise"
	*destinations[2].(*uuid.UUID) = uuid.MustParse("40000000-0000-4000-8000-000000000001")
	*destinations[3].(*int64) = 7
}
