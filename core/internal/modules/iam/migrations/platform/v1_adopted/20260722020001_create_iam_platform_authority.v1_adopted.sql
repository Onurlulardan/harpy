-- +goose Up
CREATE TABLE platform.iam_platform_subjects (
    id uuid NOT NULL,
    environment_id uuid NOT NULL,
    primary_tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    state text NOT NULL CHECK (state IN ('active', 'retired')),
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    retired_at timestamptz,
    CONSTRAINT iam_platform_subjects_pk PRIMARY KEY (id),
    CONSTRAINT iam_platform_subjects_environment_id_uq UNIQUE (environment_id, id),
    CONSTRAINT iam_platform_subjects_identity_uq
        UNIQUE (environment_id, primary_tenant_id, user_id),
    CONSTRAINT iam_platform_subjects_primary_binding_fk
        FOREIGN KEY (environment_id, primary_tenant_id)
        REFERENCES platform.tenancy_tenant_bindings (environment_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_subjects_retired_ck CHECK (
        (state = 'active' AND retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND retired_at >= updated_at)
    )
);

CREATE TABLE platform.iam_platform_roles (
    id uuid NOT NULL,
    environment_id uuid NOT NULL,
    role_kind text NOT NULL CHECK (role_kind IN ('instance_owner', 'platform_operator')),
    role_key text COLLATE "C" NOT NULL,
    state text NOT NULL CHECK (state IN ('active', 'retired')),
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    retired_at timestamptz,
    CONSTRAINT iam_platform_roles_pk PRIMARY KEY (id),
    CONSTRAINT iam_platform_roles_environment_id_uq UNIQUE (environment_id, id),
    CONSTRAINT iam_platform_roles_key_uq UNIQUE (environment_id, role_key),
    CONSTRAINT iam_platform_roles_policy_fk FOREIGN KEY (environment_id)
        REFERENCES platform.tenancy_deployment_policy (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_roles_key_ck CHECK (
        role_key ~ '^[a-z][a-z0-9_.-]{0,99}$'
    ),
    CONSTRAINT iam_platform_roles_retired_ck CHECK (
        (state = 'active' AND retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND retired_at >= updated_at)
    )
);

CREATE UNIQUE INDEX iam_platform_roles_owner_uq
ON platform.iam_platform_roles (environment_id)
WHERE role_kind = 'instance_owner' AND state = 'active';

CREATE TABLE platform.iam_platform_role_permissions (
    id uuid NOT NULL,
    environment_id uuid NOT NULL,
    role_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    state text NOT NULL CHECK (state IN ('active', 'revoked')),
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    revoked_at timestamptz,
    CONSTRAINT iam_platform_role_permissions_pk PRIMARY KEY (id),
    CONSTRAINT iam_platform_role_permissions_environment_id_uq
        UNIQUE (environment_id, id),
    CONSTRAINT iam_platform_role_permissions_role_fk
        FOREIGN KEY (environment_id, role_id)
        REFERENCES platform.iam_platform_roles (environment_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_role_permissions_permission_fk
        FOREIGN KEY (permission_id)
        REFERENCES platform.iam_platform_permissions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_role_permissions_revoked_ck CHECK (
        (state = 'active' AND revoked_at IS NULL) OR
        (state = 'revoked' AND revoked_at IS NOT NULL AND revoked_at >= updated_at)
    )
);

CREATE UNIQUE INDEX iam_platform_role_permissions_active_uq
ON platform.iam_platform_role_permissions (environment_id, role_id, permission_id)
WHERE state = 'active';

CREATE TABLE platform.iam_platform_subject_roles (
    id uuid NOT NULL,
    environment_id uuid NOT NULL,
    subject_id uuid NOT NULL,
    role_id uuid NOT NULL,
    state text NOT NULL CHECK (state IN ('active', 'revoked')),
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    revoked_at timestamptz,
    CONSTRAINT iam_platform_subject_roles_pk PRIMARY KEY (id),
    CONSTRAINT iam_platform_subject_roles_environment_id_uq UNIQUE (environment_id, id),
    CONSTRAINT iam_platform_subject_roles_subject_fk
        FOREIGN KEY (environment_id, subject_id)
        REFERENCES platform.iam_platform_subjects (environment_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_subject_roles_role_fk
        FOREIGN KEY (environment_id, role_id)
        REFERENCES platform.iam_platform_roles (environment_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_subject_roles_revoked_ck CHECK (
        (state = 'active' AND revoked_at IS NULL) OR
        (state = 'revoked' AND revoked_at IS NOT NULL AND revoked_at >= updated_at)
    )
);

CREATE UNIQUE INDEX iam_platform_subject_roles_active_uq
ON platform.iam_platform_subject_roles (environment_id, subject_id, role_id)
WHERE state = 'active';
