-- +goose Up
CREATE TABLE analytics_publication_releases (
    id uuid NOT NULL,
    resource_kind text NOT NULL,
    resource_version_id uuid NOT NULL,
    audience text NOT NULL,
    lifecycle_state text NOT NULL,
    cache_policy text NOT NULL,
    manifest_sha256 bytea NOT NULL,
    requester_id uuid NOT NULL,
    approver_id uuid,
    row_revision bigint NOT NULL,
    generation bigint NOT NULL,
    public_id bytea,
    capability_expires_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    submitted_at timestamptz,
    approved_at timestamptz,
    activated_at timestamptz,
    revoked_at timestamptz,
    CONSTRAINT analytics_publication_releases_pkey PRIMARY KEY (id),
    CONSTRAINT analytics_publication_releases_public_key UNIQUE (public_id),
    CONSTRAINT analytics_publication_releases_identity_check CHECK (
        pg_catalog.uuid_extract_version(id) = 7
        AND pg_catalog.uuid_extract_version(resource_version_id) = 7
        AND resource_kind IN ('query', 'dashboard')
        AND audience IN ('tenant', 'public', 'capability')
        AND lifecycle_state IN ('draft', 'pending_approval', 'active', 'revoked', 'retired')
        AND cache_policy IN ('none', 'revision_bound')
        AND pg_catalog.octet_length(manifest_sha256) = 32
        AND requester_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND (approver_id IS NULL OR approver_id <> requester_id)
        AND row_revision > 0 AND generation > 0
        AND (public_id IS NULL OR pg_catalog.octet_length(public_id) = 16)
    ),
    CONSTRAINT analytics_publication_releases_state_check CHECK (
        (lifecycle_state = 'draft' AND submitted_at IS NULL AND approver_id IS NULL)
        OR (lifecycle_state = 'pending_approval' AND submitted_at IS NOT NULL)
        OR (lifecycle_state = 'active' AND submitted_at IS NOT NULL
            AND activated_at IS NOT NULL
            AND (audience = 'tenant' OR approver_id IS NOT NULL))
        OR (lifecycle_state IN ('revoked', 'retired') AND activated_at IS NOT NULL
            AND revoked_at IS NOT NULL)
    ),
    CONSTRAINT analytics_publication_releases_public_check CHECK (
        (audience = 'tenant' AND public_id IS NULL AND capability_expires_at IS NULL)
        OR (audience = 'public' AND
            (lifecycle_state NOT IN ('active', 'revoked', 'retired') OR public_id IS NOT NULL)
            AND capability_expires_at IS NULL)
        OR (audience = 'capability'
            AND (lifecycle_state NOT IN ('active', 'revoked', 'retired')
                 OR (public_id IS NOT NULL AND capability_expires_at IS NOT NULL)))
    ),
    CONSTRAINT analytics_publication_releases_time_check CHECK (
        pg_catalog.isfinite(created_at) AND pg_catalog.isfinite(updated_at)
        AND created_at <= updated_at
        AND (submitted_at IS NULL OR pg_catalog.isfinite(submitted_at))
        AND (approved_at IS NULL OR pg_catalog.isfinite(approved_at))
        AND (activated_at IS NULL OR pg_catalog.isfinite(activated_at))
        AND (revoked_at IS NULL OR pg_catalog.isfinite(revoked_at))
        AND (capability_expires_at IS NULL OR pg_catalog.isfinite(capability_expires_at))
    )
);

CREATE TABLE analytics_publication_approvals (
    release_id uuid NOT NULL,
    requester_id uuid NOT NULL,
    requester_step_up_at timestamptz NOT NULL,
    approver_id uuid NOT NULL,
    approver_step_up_at timestamptz NOT NULL,
    approved_at timestamptz NOT NULL,
    evidence_sha256 bytea NOT NULL,
    CONSTRAINT analytics_publication_approvals_pkey PRIMARY KEY (release_id),
    CONSTRAINT analytics_publication_approvals_release_fk FOREIGN KEY (release_id)
        REFERENCES analytics_publication_releases (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_publication_approvals_shape_check CHECK (
        requester_id <> approver_id
        AND pg_catalog.isfinite(requester_step_up_at)
        AND pg_catalog.isfinite(approver_step_up_at)
        AND pg_catalog.isfinite(approved_at)
        AND requester_step_up_at <= approved_at
        AND approver_step_up_at <= approved_at
        AND approved_at - requester_step_up_at <= interval '10 minutes'
        AND approved_at - approver_step_up_at <= interval '10 minutes'
        AND pg_catalog.octet_length(evidence_sha256) = 32
    )
);

CREATE TABLE analytics_capability_tokens (
    release_id uuid NOT NULL,
    token_sha256 bytea NOT NULL,
    issued_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT analytics_capability_tokens_pkey PRIMARY KEY (release_id),
    CONSTRAINT analytics_capability_tokens_hash_key UNIQUE (token_sha256),
    CONSTRAINT analytics_capability_tokens_release_fk FOREIGN KEY (release_id)
        REFERENCES analytics_publication_releases (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_capability_tokens_shape_check CHECK (
        pg_catalog.octet_length(token_sha256) = 32
        AND pg_catalog.isfinite(issued_at) AND pg_catalog.isfinite(expires_at)
        AND expires_at > issued_at AND expires_at - issued_at <= interval '30 days'
        AND (revoked_at IS NULL OR
             (pg_catalog.isfinite(revoked_at) AND revoked_at >= issued_at))
    )
);

CREATE TABLE analytics_execution_audits (
    execution_id uuid NOT NULL,
    release_id uuid NOT NULL,
    principal_class text NOT NULL,
    authorization_sha256 bytea NOT NULL,
    parameter_sha256 bytea NOT NULL,
    result_sha256 bytea NOT NULL,
    result_bytes integer NOT NULL,
    cache_hit boolean NOT NULL,
    occurred_at timestamptz NOT NULL,
    CONSTRAINT analytics_execution_audits_pkey PRIMARY KEY (execution_id),
    CONSTRAINT analytics_execution_audits_release_fk FOREIGN KEY (release_id)
        REFERENCES analytics_publication_releases (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_execution_audits_shape_check CHECK (
        pg_catalog.uuid_extract_version(execution_id) = 7
        AND principal_class IN ('actor', 'public', 'capability')
        AND pg_catalog.octet_length(authorization_sha256) = 32
        AND pg_catalog.octet_length(parameter_sha256) = 32
        AND pg_catalog.octet_length(result_sha256) = 32
        AND result_bytes BETWEEN 0 AND 1048576
        AND pg_catalog.isfinite(occurred_at)
    )
);

CREATE INDEX analytics_publication_releases_list_idx
    ON analytics_publication_releases (id);
CREATE INDEX analytics_execution_audits_retention_idx
    ON analytics_execution_audits (occurred_at, execution_id);

