//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

const (
	iamInvitationManager    = "51000000-0000-4000-8000-000000000001"
	iamInvitationAcceptUser = "51000000-0000-4000-8000-000000000003"
	iamInvitationIdentifier = "b0010000-0000-4000-8000-000000000001"
	iamInvitationRole       = "b0040000-0000-4000-8000-000000000001"
)

type iamInvitationRuntimeFixture struct {
	id           string
	tokenID      string
	assignmentID string
	operationID  string
	selector     []byte
	hmac         []byte
	binding      []byte
}

func assertIAMInvitationRuntime(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	marker string,
) {
	t.Helper()
	accepted := createIAMInvitation(
		t, ctx, database, tenant, marker, "c1", "c2", "invitee@example.com",
	)
	revoked := createIAMInvitation(
		t, ctx, database, tenant, marker, "c5", "c6", "revoked-"+marker+"@example.com",
	)
	assertIAMInvitationReads(t, ctx, database, tenant.schema, accepted.id)
	revokeIAMInvitation(t, ctx, database, tenant.schema, revoked.id)
	expireIAMInvitation(t, ctx, database, tenant.schema)
	acceptIAMInvitation(t, ctx, database, tenant, marker, accepted)
}

func createIAMInvitation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	marker string,
	invitationPrefix string,
	tokenPrefix string,
	email string,
) iamInvitationRuntimeFixture {
	t.Helper()
	fixture := iamInvitationRuntimeFixture{
		id:           iamPermissionInvitationID(invitationPrefix, marker),
		tokenID:      iamPermissionInvitationID(tokenPrefix, marker),
		assignmentID: iamPermissionInvitationID("c3", marker),
		operationID:  iamPermissionInvitationID("c4", marker),
		selector:     []byte("sel-" + invitationPrefix + "-" + marker + "-000000"),
		hmac:         []byte(strings.Repeat(invitationPrefix[1:], 32)),
		binding:      []byte(strings.Repeat(tokenPrefix[1:], 32)),
	}
	result := mustRunIAMPGFirstText(t, ctx, database, tenant.schema, `
SELECT pg_catalog.concat_ws('|',invitation_id::text,invitation_state,
    invitation_revision::text)
FROM iam_invitation_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5,'tenant',$6,NULL,$7,1,$8,$9,'one-time-1',$10,$11
)`, fixture.id, email, email, []byte(strings.Repeat("n", 32)),
		iamInvitationRole, tenant.id, iamInvitationManager, fixture.tokenID,
		fixture.selector, fixture.hmac, fixture.binding)
	if result != fixture.id+"|pending|1" {
		t.Fatalf("invitation create = %q", result)
	}
	return fixture
}

func assertIAMInvitationReads(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	invitationID string,
) {
	t.Helper()
	read := mustRunIAMPGFirstText(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|',invitation_id::text,invitation_state,
    invitation_revision::text)
FROM iam_invitation_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2
)`, iamInvitationManager, invitationID)
	if read != invitationID+"|pending|1" {
		t.Fatalf("invitation read = %q", read)
	}
	listed := mustRunIAMPGFirstText(t, ctx, database, schema, `
SELECT pg_catalog.count(*)::text FROM iam_invitation_list_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,'["pending"]'::jsonb,NULL,101
)`, iamInvitationManager)
	if listed != "3" {
		t.Fatalf("pending invitation count = %s, want 3", listed)
	}
}

func revokeIAMInvitation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	invitationID string,
) {
	t.Helper()
	result := mustRunIAMPGFirstText(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|',invitation_id::text,invitation_state,
    invitation_revision::text)
FROM iam_invitation_revoke_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,1,$2,1
)`, invitationID, iamInvitationManager)
	if result != invitationID+"|revoked|2" {
		t.Fatalf("invitation revoke = %q", result)
	}
}

func expireIAMInvitation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	const invitationID = "be000000-0000-4000-8000-000000000001"
	result := mustRunIAMPGFirstText(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|',invitation_id::text,invitation_state,
    invitation_revision::text)
FROM iam_invitation_expire_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,1,$2,1
)`, invitationID, iamInvitationManager)
	if result != invitationID+"|expired|2" {
		t.Fatalf("invitation expire = %q", result)
	}
}
