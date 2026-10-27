package postgres

import (
	"bytes"
	"testing"
)

var iamInvitationTokenMigrationNames = []string{
	"20260716012720_create_iam_invitations",
	"20260716012750_create_iam_one_time_tokens",
}

func TestIAMInvitationTokenMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamInvitationTokenMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and V1-adopted invitation/token migration differs")
			}
			assertIAMIdentityMigrationSource(t, fresh)
		})
	}
}

func TestIAMInvitationTokenMigrationContracts(t *testing.T) {
	tests := []struct {
		name        string
		foreignKeys int
		markers     []string
	}{
		{
			name:        iamInvitationTokenMigrationNames[0],
			foreignKeys: 4,
			markers: []string{
				"CREATE TABLE iam_invitations (",
				"identifier_kind = 'email_v1'",
				"normalized_value text COLLATE \"C\" NOT NULL",
				"REFERENCES iam_roles (id, scope_kind, scope_id)",
				"REFERENCES iam_organizations (id)",
				"REFERENCES iam_users (id)",
				"REFERENCES iam_user_identifiers (",
				"id, user_id, kind, normalized_value",
				"role_scope_kind = 'organization'",
				"state IN ('pending', 'accepted', 'revoked', 'expired')",
				"expires_at = created_at + INTERVAL '7 days'",
				"iam_invitations_one_pending_target_uidx",
			},
		},
		{
			name:        iamInvitationTokenMigrationNames[1],
			foreignKeys: 4,
			markers: []string{
				"CREATE TABLE iam_one_time_tokens (",
				"UNIQUE (selector)",
				"UNIQUE (hash_key_id, secret_hmac)",
				"REFERENCES iam_users (id)",
				"FOREIGN KEY (identifier_id, user_id, identifier_kind)",
				"REFERENCES iam_user_identifiers (id, user_id, kind)",
				"REFERENCES iam_invitations (id, expires_at)",
				"REFERENCES iam_portal_subject_bindings (id, user_id)",
				"'portal_setup_v1', 'invitation_accept_v1'",
				"identifier_kind = 'email_v1'",
				"identifier_kind = 'portal_v1'",
				"pg_catalog.octet_length(selector) = 16",
				"pg_catalog.octet_length(secret_hmac) = 32",
				"pg_catalog.octet_length(binding_sha256) = 32",
				"attempt_count BETWEEN 0 AND 5",
				"expires_at = created_at + INTERVAL '1 hour'",
				"expires_at = created_at + INTERVAL '24 hours'",
				"expires_at = invitation_expires_at AND created_at < expires_at",
				"iam_one_time_tokens_one_pending_binding_uidx",
				"iam_one_time_tokens_one_pending_identifier_uidx",
				"iam_one_time_tokens_one_pending_invitation_uidx",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := string(readIAMIdentityMigration(t, "fresh", test.name))
			assertIAMIdentityMarkers(t, source, test.foreignKeys, test.markers)
		})
	}
}
