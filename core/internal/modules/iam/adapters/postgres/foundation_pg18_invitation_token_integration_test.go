//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMInvitationTokenConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		mustExecIAMTenant(
			t, ctx, database, tenant.schema,
			insertIAMOTPRole, iamOTPRoleID, iamOTPRoleScopeID,
		)
		mustExecIAMTenant(
			t, ctx, database, tenant.schema, insertIAMActiveIdentifier,
			iamOTPIdentifierID, iamUserTwoID,
			"OTP.User@example.com", "otp.user@example.com",
		)
		mustExecIAMTenant(
			t, ctx, database, tenant.schema, insertIAMOTPPortalIdentifier,
			iamOTPPortalID, iamUserTwoID,
		)

		err := rejectIAMTenant(
			t, ctx, database, tenant.schema, insertIAMInvitationFixture,
			"87000000-0000-4000-8000-000000000099",
			"Wrong Scope@example.com", "wrong.scope@example.com", "89",
			iamOTPRoleID, "85000000-0000-4000-8000-000000000099", iamUserOneID,
		)
		assertIAMConstraint(t, err, "23503", "iam_invitations_role_fk")

		mustExecIAMTenant(
			t, ctx, database, tenant.schema, insertIAMInvitationFixture,
			iamOTPInvitationID, "Invitee@example.com", "invitee@example.com",
			"88", iamOTPRoleID, iamOTPRoleScopeID, iamUserOneID,
		)
		positive := []struct {
			id, purpose, user, identifier, kind, invitation string
			selector, hmac, binding, ttl                    string
		}{
			{"88000000-0000-4000-8000-000000000001", "password_reset_v1",
				iamUserTwoID, iamOTPIdentifierID, "email_v1", "", "91", "a1", "b1", "1 hour"},
			{"88000000-0000-4000-8000-000000000002", "email_verification_v1",
				iamUserTwoID, iamOTPIdentifierID, "email_v1", "", "92", "a2", "b2", "24 hours"},
			{"88000000-0000-4000-8000-000000000003", "portal_setup_v1",
				iamUserTwoID, iamOTPPortalID, "portal_v1", "", "93", "a3", "b3", "1 hour"},
			{"88000000-0000-4000-8000-000000000004", "invitation_accept_v1",
				"", "", "", iamOTPInvitationID, "94", "a4", "b4", "7 days"},
		}
		for _, token := range positive {
			mustExecIAMTenant(
				t, ctx, database, tenant.schema, insertIAMOneTimeTokenFixture,
				token.id, token.purpose, nilIfEmpty(token.user),
				nilIfEmpty(token.identifier), nilIfEmpty(token.kind),
				nilIfEmpty(token.invitation), token.selector, token.hmac,
				token.binding, token.ttl,
			)
		}
		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_one_time_tokens
SET portal_binding_id = '53000000-0000-4000-8000-000000000001'
WHERE id = '88000000-0000-4000-8000-000000000003'`)
		assertIAMConstraint(t, err, "23503", "iam_one_time_tokens_portal_binding_fk")
		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_one_time_tokens SET portal_setup_revision = NULL
WHERE id = '88000000-0000-4000-8000-000000000003'`)
		assertIAMConstraint(t, err, "23514", "iam_one_time_tokens_subject_shape_check")
		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_one_time_tokens SET identifier_revision = NULL
WHERE id = '88000000-0000-4000-8000-000000000001'`)
		assertIAMConstraint(t, err, "23514", "iam_one_time_tokens_subject_shape_check")

		err = rejectIAMTenant(
			t, ctx, database, tenant.schema, insertIAMOneTimeTokenFixture,
			"88000000-0000-4000-8000-000000000091", "password_reset_v1",
			iamUserOneID, iamOTPIdentifierID, "email_v1", nil,
			"d1", "e1", "f1", "1 hour",
		)
		assertIAMConstraint(t, err, "23503", "iam_one_time_tokens_identifier_user_fk")

		err = rejectIAMTenant(
			t, ctx, database, tenant.schema, insertIAMOneTimeTokenFixture,
			"88000000-0000-4000-8000-000000000092", "email_verification_v1",
			iamUserTwoID, iamOTPPortalID, "portal_v1", nil,
			"d2", "e2", "f2", "24 hours",
		)
		assertIAMConstraint(t, err, "23514", "iam_one_time_tokens_subject_shape_check")

		err = rejectIAMTenant(
			t, ctx, database, tenant.schema, insertIAMOneTimeTokenFixture,
			"88000000-0000-4000-8000-000000000093", "password_reset_v1",
			iamUserTwoID, iamOTPIdentifierID, "email_v1", nil,
			"d3", "e3", "b1", "1 hour",
		)
		assertIAMConstraint(t, err, "23505", "iam_one_time_tokens_one_pending_binding_uidx")
		err = rejectIAMTenant(
			t, ctx, database, tenant.schema, insertIAMOneTimeTokenFixture,
			"88000000-0000-4000-8000-000000000094", "password_reset_v1",
			iamUserTwoID, iamOTPIdentifierID, "email_v1", nil,
			"d4", "e4", "f4", "1 hour",
		)
		assertIAMConstraint(t, err, "23505", "iam_one_time_tokens_one_pending_identifier_uidx")

		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_invitations
SET state = 'accepted', accepted_user_id = $1,
    accepted_identifier_id = $2, accepted_identifier_kind = 'email_v1',
    accepted_at = '2026-07-16T01:00:00Z', updated_at = '2026-07-16T01:00:00Z'
WHERE id = $3`, iamUserOneID, iamOTPIdentifierID, iamOTPInvitationID)
		assertIAMConstraint(t, err, "23503", "iam_invitations_accepted_identifier_fk")

		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_one_time_tokens
SET invitation_expires_at = invitation_expires_at + INTERVAL '1 second',
    expires_at = expires_at + INTERVAL '1 second',
    purge_after = purge_after + INTERVAL '1 second'
WHERE id = '88000000-0000-4000-8000-000000000004'`)
		assertIAMConstraint(t, err, "23503", "iam_one_time_tokens_invitation_fk")

		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_one_time_tokens
SET expires_at = created_at + INTERVAL '2 hours',
    purge_after = created_at + INTERVAL '26 hours'
WHERE id = '88000000-0000-4000-8000-000000000001'`)
		assertIAMConstraint(t, err, "23514", "iam_one_time_tokens_ttl_check")

		err = rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_one_time_tokens SET attempt_count = 5
WHERE id = '88000000-0000-4000-8000-000000000001'`)
		assertIAMConstraint(t, err, "23514", "iam_one_time_tokens_lifecycle_check")
		assertIAMNoPlaintextTokenColumns(t, ctx, database, tenant.schema)
	}
}

func nilIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func assertIAMNoPlaintextTokenColumns(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	var exposed int
	err := database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*)
FROM information_schema.columns
WHERE table_schema = $1 AND table_name = 'iam_one_time_tokens' AND (
    column_name IN (
        'token', 'token_value', 'secret', 'secret_value', 'plaintext_secret',
        'raw_secret', 'password', 'code'
    ) OR (
        column_name IN ('selector', 'secret_hmac', 'binding_sha256') AND
        data_type <> 'bytea'
    )
)`, schema).Scan(&exposed)
	if err != nil {
		t.Fatalf("inspect %s one-time token surface: %v", schema, err)
	}
	if exposed != 0 {
		t.Fatalf("%s one-time token plaintext/typed exposure count = %d", schema, exposed)
	}
}
