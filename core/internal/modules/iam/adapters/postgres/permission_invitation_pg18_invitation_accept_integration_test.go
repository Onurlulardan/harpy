//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func acceptIAMInvitation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	marker string,
	fixture iamInvitationRuntimeFixture,
) {
	t.Helper()
	arguments := []any{
		fixture.operationID, []byte(strings.Repeat("q", 32)), fixture.id,
		fixture.tokenID, fixture.selector, fixture.hmac, fixture.binding,
		iamInvitationAcceptUser, iamInvitationIdentifier, fixture.assignmentID,
	}
	badArguments := append([]any(nil), arguments...)
	badArguments[5] = []byte(strings.Repeat("x", 32))
	err := rejectIAMPGFirst(
		t, ctx, database, tenant.schema, iamInvitationAcceptSQL, badArguments...,
	)
	assertIAMSQLState(t, err, "42501")
	rollbackIAMPGFirstText(
		t, ctx, database, tenant.schema, iamInvitationAcceptSQL, arguments...,
	)
	assertInvitationAcceptState(
		t, ctx, database, tenant.schema, fixture, "pending", "pending", 0, 1, 1,
	)
	result := mustRunIAMPGFirstText(
		t, ctx, database, tenant.schema, iamInvitationAcceptSQL, arguments...,
	)
	if result != "iam.invitation.accepted|2|2|false" {
		t.Fatalf("invitation accept = %q", result)
	}
	assertInvitationAcceptState(
		t, ctx, database, tenant.schema, fixture, "accepted", "consumed", 1, 2, 2,
	)
	replayed := mustRunIAMPGFirstText(
		t, ctx, database, tenant.schema, iamInvitationAcceptSQL, arguments...,
	)
	if replayed != "iam.invitation.accepted|2|2|true" {
		t.Fatalf("invitation accept replay = %q", replayed)
	}
	staleArguments := append([]any(nil), arguments...)
	staleArguments[0] = iamPermissionInvitationID("c9", marker)
	err = rejectIAMPGFirst(
		t, ctx, database, tenant.schema, iamInvitationAcceptSQL, staleArguments...,
	)
	assertIAMSQLState(t, err, "40001")
}

const iamInvitationAcceptSQL = `
SELECT pg_catalog.concat_ws('|',result_code,result_revision::text,
    authorization_clock::text,replayed::text)
FROM iam_invitation_accept_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,1,$4,1,$5,'one-time-1',$6,$7,$8,$9,$10,NULL,1
)`

func assertInvitationAcceptState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	fixture iamInvitationRuntimeFixture,
	wantInvitation string,
	wantToken string,
	wantAssignments int64,
	wantUserVersion int64,
	wantClock int64,
) {
	t.Helper()
	var invitationState, tokenState string
	var assignmentCount, userVersion, clock int64
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT invitation.state, token.state,
       (SELECT pg_catalog.count(*) FROM iam_user_role_assignments AS assignment
        WHERE assignment.id = '`+fixture.assignmentID+`'),
       principal.authorization_version, clock_state.clock_value
FROM iam_invitations AS invitation
JOIN iam_one_time_tokens AS token
  ON token.id = '`+fixture.tokenID+`' AND token.invitation_id = invitation.id
JOIN iam_users AS principal
  ON principal.id = '51000000-0000-4000-8000-000000000003'
CROSS JOIN iam_authorization_clock AS clock_state
WHERE invitation.id = '`+fixture.id+`' AND clock_state.singleton`,
		&invitationState, &tokenState, &assignmentCount, &userVersion, &clock)
	if invitationState != wantInvitation || tokenState != wantToken ||
		assignmentCount != wantAssignments || userVersion != wantUserVersion ||
		clock != wantClock {
		t.Fatalf(
			"invitation state = %s/%s assignments=%d user=%d clock=%d",
			invitationState, tokenState, assignmentCount, userVersion, clock,
		)
	}
}
