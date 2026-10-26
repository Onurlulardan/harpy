package postgres

import "testing"

func TestIAMHumanAccessMigrationContracts(t *testing.T) {
	tests := []struct {
		name        string
		foreignKeys int
		markers     []string
	}{
		{
			name:        "20260716012100_correct_iam_human_access",
			foreignKeys: 1,
			markers: []string{
				"ADD COLUMN identity_kind text",
				"SET identity_kind = 'human'",
				"CHECK (identity_kind = 'human')",
				"CREATE TABLE iam_user_access_modes (",
				"PRIMARY KEY (id)",
				"UNIQUE (user_id, mode)",
				"UNIQUE (id, user_id, mode)",
				"REFERENCES iam_users (id)",
				"mode IN ('operator_console', 'portal')",
				"state IN ('pending', 'active', 'suspended', 'retired')",
				"DROP CONSTRAINT iam_users_account_type_check",
				"DROP COLUMN account_type",
			},
		},
		{
			name:        "20260716012120_create_iam_portal_subject_bindings",
			foreignKeys: 1,
			markers: []string{
				"CREATE TABLE iam_portal_subject_bindings (",
				"UNIQUE (id, user_id, access_mode_id)",
				"REFERENCES iam_user_access_modes (id, user_id, mode)",
				"CHECK (access_mode = 'portal')",
				"subject_shape IN ('self', 'form_record')",
				"subject_type = 'operator'",
				"subject_type ~ '^[a-z0-9_-]+$'",
				"profile_owner = 'iam'",
				"profile_owner = 'forms'",
				"pg_catalog.octet_length(contract_sha256) = 32",
				"iam_portal_subject_bindings_live_user_uidx",
				"iam_portal_subject_bindings_live_form_record_uidx",
			},
		},
		{
			name:        "20260716012130_create_iam_portal_access_setups",
			foreignKeys: 3,
			markers: []string{
				"CREATE TABLE iam_portal_access_setups (",
				"PRIMARY KEY (binding_id)",
				"REFERENCES iam_portal_subject_bindings (id, user_id, access_mode_id)",
				"REFERENCES iam_user_access_modes (id, user_id, mode)",
				"REFERENCES iam_users (id)",
				"setup_state IN ('not_required', 'required', 'completed')",
				"'revision_required', 'approved', 'rejected'",
				"provenance IN ('native_v2', 'v1_adopted')",
				"iam_portal_access_setups_state_combination_check",
				"AND approval_revision > 0",
				"pg_catalog.octet_length(source_evidence_sha256) = 32",
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
