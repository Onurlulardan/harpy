-- +goose Up
CREATE TABLE platform.iam_platform_resources (
    id uuid PRIMARY KEY,
    resource_key text COLLATE "C" NOT NULL UNIQUE,
    state text NOT NULL CHECK (state IN ('active', 'retired')),
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    retired_at timestamptz,
    CONSTRAINT iam_platform_resources_key_ck CHECK (
        resource_key ~ '^platform(\.[a-z][a-z0-9_]*([.-][a-z0-9_]+)*)+$'
    ),
    CONSTRAINT iam_platform_resources_retired_ck CHECK (
        (state = 'active' AND retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND retired_at >= updated_at)
    )
);

CREATE TABLE platform.iam_platform_actions (
    id uuid PRIMARY KEY,
    action_key text COLLATE "C" NOT NULL UNIQUE,
    state text NOT NULL CHECK (state IN ('active', 'retired')),
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    retired_at timestamptz,
    CONSTRAINT iam_platform_actions_key_ck CHECK (
        action_key ~ '^[a-z][a-z0-9_]*([.-][a-z0-9_]+)*$'
    ),
    CONSTRAINT iam_platform_actions_retired_ck CHECK (
        (state = 'active' AND retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND retired_at >= updated_at)
    )
);

CREATE TABLE platform.iam_platform_permissions (
    id uuid PRIMARY KEY,
    permission_key text COLLATE "C" NOT NULL UNIQUE,
    match_kind text NOT NULL CHECK (match_kind IN ('exact', 'environment_all')),
    resource_id uuid,
    action_id uuid,
    deployment_mode text NOT NULL CHECK (deployment_mode IN ('both', 'onpremise', 'saas')),
    delegable boolean NOT NULL,
    step_up_required boolean NOT NULL,
    owner_only boolean NOT NULL,
    state text NOT NULL CHECK (state IN ('active', 'retired')),
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    retired_at timestamptz,
    CONSTRAINT iam_platform_permissions_resource_fk FOREIGN KEY (resource_id)
        REFERENCES platform.iam_platform_resources (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_permissions_action_fk FOREIGN KEY (action_id)
        REFERENCES platform.iam_platform_actions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_platform_permissions_shape_ck CHECK (
        (match_kind = 'exact' AND resource_id IS NOT NULL AND action_id IS NOT NULL AND
         permission_key ~ '^platform\.') OR
        (match_kind = 'environment_all' AND resource_id IS NULL AND action_id IS NULL AND
         permission_key = 'environment_all')
    ),
    CONSTRAINT iam_platform_permissions_delegation_ck CHECK (
        NOT owner_only OR NOT delegable
    ),
    CONSTRAINT iam_platform_permissions_retired_ck CHECK (
        (state = 'active' AND retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND retired_at >= updated_at)
    ),
    UNIQUE NULLS NOT DISTINCT (resource_id, action_id)
);
