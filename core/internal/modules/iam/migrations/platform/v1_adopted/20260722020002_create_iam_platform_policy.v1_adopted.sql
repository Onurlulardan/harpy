-- +goose Up
CREATE TABLE platform.iam_platform_owner_policies (
    environment_id uuid PRIMARY KEY,
    owner_binding_id uuid NOT NULL UNIQUE,
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    CONSTRAINT iam_platform_owner_policies_binding_fk
        FOREIGN KEY (environment_id, owner_binding_id)
        REFERENCES platform.iam_platform_subject_roles (environment_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_owner_policies_environment_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.tenancy_deployment_policy (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE platform.iam_platform_authorization_clocks (
    environment_id uuid PRIMARY KEY,
    authorization_clock bigint NOT NULL CHECK (authorization_clock > 0),
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    CONSTRAINT iam_platform_authorization_clocks_environment_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.iam_platform_owner_policies (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE platform.iam_platform_mutation_receipts (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    mutation_kind text NOT NULL CHECK (
        mutation_kind IN ('initial_owner', 'operator_grant', 'operator_revoke')
    ),
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256) = 32),
    actor_user_id uuid NOT NULL,
    target_user_id uuid NOT NULL,
    subject_id uuid NOT NULL,
    role_id uuid NOT NULL,
    subject_role_binding_id uuid NOT NULL,
    role_permission_binding_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    target_authorization_version bigint NOT NULL CHECK (target_authorization_version > 0),
    platform_clock bigint NOT NULL CHECK (platform_clock > 0),
    occurred_at timestamptz NOT NULL,
    CONSTRAINT iam_platform_receipts_policy_fk FOREIGN KEY (environment_id)
        REFERENCES platform.iam_platform_owner_policies (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_receipts_subject_fk FOREIGN KEY (environment_id, subject_id)
        REFERENCES platform.iam_platform_subjects (environment_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_receipts_role_fk FOREIGN KEY (environment_id, role_id)
        REFERENCES platform.iam_platform_roles (environment_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_receipts_subject_role_fk
        FOREIGN KEY (environment_id, subject_role_binding_id)
        REFERENCES platform.iam_platform_subject_roles (environment_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_receipts_role_permission_fk
        FOREIGN KEY (environment_id, role_permission_binding_id)
        REFERENCES platform.iam_platform_role_permissions (environment_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_receipts_permission_fk FOREIGN KEY (permission_id)
        REFERENCES platform.iam_platform_permissions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE INDEX iam_platform_receipts_target_idx
ON platform.iam_platform_mutation_receipts (environment_id, target_user_id, occurred_at);
