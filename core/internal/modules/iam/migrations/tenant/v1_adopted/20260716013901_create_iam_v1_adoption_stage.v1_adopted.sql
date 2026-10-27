-- +goose Up
CREATE TABLE iam_v1_adoption_stage (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    source_category text NOT NULL,
    source_key text NOT NULL,
    target_table text NOT NULL,
    target_key text NOT NULL,
    target_record jsonb NOT NULL,
    record_sha256 bytea NOT NULL,
    removed boolean NOT NULL,
    source_xid xid8 NOT NULL,
    batch_ordinal integer NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_v1_adoption_stage_pkey
        PRIMARY KEY (operation_id, run_id, target_table, target_key),
    CONSTRAINT iam_v1_adoption_stage_category_fk
        FOREIGN KEY (operation_id, run_id, source_category)
        REFERENCES iam_v1_adoption_category_evidence (operation_id, run_id, category),
    CONSTRAINT iam_v1_adoption_stage_table_check CHECK (
        target_table IN (
            'iam_profile_schema_versions', 'iam_profile_fields',
            'iam_profile_field_revisions', 'iam_organizations', 'iam_users',
            'iam_user_identifiers', 'iam_password_credentials',
            'iam_organization_memberships', 'iam_roles',
            'iam_user_role_assignments', 'iam_user_access_modes',
            'iam_portal_subject_bindings', 'iam_portal_access_setups',
            'iam_resources', 'iam_actions', 'iam_permissions',
            'iam_user_permission_bindings', 'iam_role_permission_bindings',
            'iam_organization_permission_bindings', 'iam_authorization_clock',
            'iam_protected_admin_policies', 'iam_permission_request_groups',
            'iam_permission_requests', 'iam_invitations', 'iam_one_time_tokens'
        )
    ),
    CONSTRAINT iam_v1_adoption_stage_key_check CHECK (
        pg_catalog.char_length(source_key) BETWEEN 1 AND 512 AND
        pg_catalog.char_length(target_key) BETWEEN 1 AND 512
    ),
    CONSTRAINT iam_v1_adoption_stage_record_check CHECK (
        pg_catalog.jsonb_typeof(target_record) = 'object' AND
        target_record <> '{}'::jsonb AND
        pg_catalog.octet_length(target_record::text) <= 1048576 AND
        pg_catalog.octet_length(record_sha256) = 32
    ),
    CONSTRAINT iam_v1_adoption_stage_batch_check CHECK (batch_ordinal > 0)
);

CREATE INDEX iam_v1_adoption_stage_source_idx
    ON iam_v1_adoption_stage (operation_id, run_id, source_category, source_key);
CREATE INDEX iam_v1_adoption_stage_live_target_idx
    ON iam_v1_adoption_stage (operation_id, run_id, target_table, target_key)
    WHERE NOT removed;
