package postgres

import (
	"testing"
)

var iamSessionMigrationNames = []string{
	"20260716012200_create_iam_sessions",
	"20260716012800_create_iam_refresh_families",
	"20260716012900_create_iam_refresh_tokens",
}

func TestIAMSessionStorageMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamSessionMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			assertIAMIdentityMigrationSource(t, fresh)
		})
	}
}

func TestIAMSessionStorageMigrationContracts(t *testing.T) {
	tests := []struct {
		name        string
		foreignKeys int
		markers     []string
	}{
		{
			name:        iamSessionMigrationNames[0],
			foreignKeys: 3,
			markers: []string{
				"CREATE TABLE iam_sessions (",
				"UNIQUE (id, user_id)",
				"REFERENCES iam_users (id)",
				"REFERENCES iam_user_access_modes (id, user_id, mode)",
				"REFERENCES iam_portal_subject_bindings (id, user_id, access_mode_id)",
				"access_mode = 'operator_console' AND portal_binding_id IS NULL",
				"access_mode = 'portal' AND portal_binding_id IS NOT NULL",
				"state IN ('active', 'revoked', 'expired')",
				"pg_catalog.octet_length(ip_hmac) = 32",
				"INTERVAL '30 days'",
			},
		},
		{
			name:        iamSessionMigrationNames[1],
			foreignKeys: 2,
			markers: []string{
				"CREATE TABLE iam_refresh_families (",
				"UNIQUE (session_id)",
				"UNIQUE (id, absolute_expires_at)",
				"REFERENCES iam_users (id)",
				"REFERENCES iam_sessions (id, user_id)",
				"state IN ('active', 'revoked', 'compromised', 'expired')",
				"absolute_expires_at = created_at + INTERVAL '90 days'",
				"retain_until >= absolute_expires_at + INTERVAL '24 hours'",
			},
		},
		{
			name:        iamSessionMigrationNames[2],
			foreignKeys: 2,
			markers: []string{
				"CREATE TABLE iam_refresh_tokens (",
				"UNIQUE (selector)",
				"UNIQUE (hash_key_id, secret_hmac)",
				"UNIQUE (family_id, sequence)",
				"UNIQUE (id, family_id, sequence)",
				"REFERENCES iam_refresh_families (id, absolute_expires_at)",
				"REFERENCES iam_refresh_tokens (id, family_id, sequence)",
				"sequence = parent_sequence + 1",
				"pg_catalog.octet_length(selector) = 16",
				"pg_catalog.octet_length(secret_hmac) = 32",
				"acr IN ('login', 'mfa', 'step_up')",
				"idle_expires_at <= issued_at + INTERVAL '30 days'",
				"iam_refresh_tokens_one_active_family_uidx",
				"iam_refresh_tokens_one_successor_uidx",
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
