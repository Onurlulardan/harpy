-- +goose Up
CREATE TABLE iam_capability_grants (
    id uuid NOT NULL,
    principal_id uuid NOT NULL,
    capability_key text COLLATE "C" NOT NULL,
    current_revision bigint NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT iam_capability_grants_pkey PRIMARY KEY (id),
    CONSTRAINT iam_capability_grants_identity_key
        UNIQUE (id, principal_id, capability_key),
    CONSTRAINT iam_capability_grants_principal_fk FOREIGN KEY (principal_id)
        REFERENCES iam_service_principals (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_capability_grants_key_check CHECK (
        pg_catalog.char_length(capability_key) BETWEEN 3 AND 100 AND
        capability_key ~ '^[a-z][a-z0-9]*([._-][a-z0-9]+)*$'
    ),
    CONSTRAINT iam_capability_grants_current_revision_check
        CHECK (current_revision > 0),
    CONSTRAINT iam_capability_grants_state_check
        CHECK (state IN ('active', 'revoked')),
    CONSTRAINT iam_capability_grants_revision_check CHECK (revision > 0),
    CONSTRAINT iam_capability_grants_timestamp_order_check CHECK (
        updated_at >= created_at AND
        (revoked_at IS NULL OR
            (revoked_at >= created_at AND revoked_at <= updated_at))
    ),
    CONSTRAINT iam_capability_grants_lifecycle_check CHECK (
        (state = 'active' AND revoked_at IS NULL) OR
        (state = 'revoked' AND revoked_at IS NOT NULL)
    )
);

CREATE TABLE iam_capability_grant_revisions (
    grant_id uuid NOT NULL,
    grant_revision bigint NOT NULL,
    previous_revision bigint,
    principal_id uuid NOT NULL,
    capability_key text COLLATE "C" NOT NULL,
    scope_kind text NOT NULL,
    owner_domain text COLLATE "C" NOT NULL,
    resource_kind text COLLATE "C" NOT NULL,
    resource_id uuid,
    valid_from timestamptz NOT NULL,
    valid_until timestamptz,
    disposition text NOT NULL,
    manifest_sha256 bytea NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_capability_grant_revisions_pkey
        PRIMARY KEY (grant_id, grant_revision),
    CONSTRAINT iam_capability_grant_revisions_head_fk
        FOREIGN KEY (grant_id, principal_id, capability_key)
        REFERENCES iam_capability_grants (id, principal_id, capability_key)
        ON UPDATE NO ACTION ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT iam_capability_grant_revisions_previous_fk
        FOREIGN KEY (grant_id, previous_revision)
        REFERENCES iam_capability_grant_revisions (grant_id, grant_revision)
        ON UPDATE NO ACTION ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT iam_capability_grant_revisions_actor_fk
        FOREIGN KEY (created_by_user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_capability_grant_revisions_chain_check CHECK (
        (grant_revision = 1 AND previous_revision IS NULL) OR
        (grant_revision > 1 AND previous_revision = grant_revision - 1)
    ),
    CONSTRAINT iam_capability_grant_revisions_scope_check CHECK (
        scope_kind IN ('tenant', 'resource') AND
        pg_catalog.char_length(owner_domain) BETWEEN 1 AND 63 AND
        owner_domain ~ '^[a-z][a-z0-9_]{0,62}$' AND
        pg_catalog.char_length(resource_kind) BETWEEN 1 AND 63 AND
        resource_kind ~ '^[a-z][a-z0-9_]{0,62}$' AND
        ((scope_kind = 'tenant' AND resource_id IS NULL) OR
            (scope_kind = 'resource' AND resource_id IS NOT NULL))
    ),
    CONSTRAINT iam_capability_grant_revisions_validity_check CHECK (
        valid_until IS NULL OR valid_until > valid_from
    ),
    CONSTRAINT iam_capability_grant_revisions_disposition_check
        CHECK (disposition IN ('active', 'revoked')),
    CONSTRAINT iam_capability_grant_revisions_manifest_check
        CHECK (
            pg_catalog.octet_length(manifest_sha256) = 32 AND
            manifest_sha256 <>
                pg_catalog.decode(pg_catalog.repeat('00', 32), 'hex')
        )
);

ALTER TABLE iam_capability_grants
    ADD CONSTRAINT iam_capability_grants_current_revision_fk
    FOREIGN KEY (id, current_revision)
    REFERENCES iam_capability_grant_revisions (grant_id, grant_revision)
    ON UPDATE NO ACTION ON DELETE RESTRICT
    DEFERRABLE INITIALLY DEFERRED;

CREATE INDEX iam_capability_grants_principal_state_idx
    ON iam_capability_grants (principal_id, state, capability_key);

CREATE INDEX iam_capability_grant_revisions_scope_idx
    ON iam_capability_grant_revisions (
        principal_id, capability_key, owner_domain, resource_kind, resource_id
    );
