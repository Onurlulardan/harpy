-- +goose Up
CREATE TABLE platform.iam_auth_throttle_buckets (
    bucket_scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    purpose text NOT NULL,
    dimension text NOT NULL,
    key_id text COLLATE "C" NOT NULL,
    bucket_hmac bytea NOT NULL,
    environment_id uuid,
    tenant_id uuid,
    authority_scope_kind text,
    authority_scope_id uuid,
    authority_group_id text,
    authority_group_generation bigint,
    authority_domain text,
    cutover_generation bigint,
    policy_version integer NOT NULL,
    window_start timestamptz NOT NULL,
    window_end timestamptz NOT NULL,
    attempt_count integer NOT NULL,
    blocked_until timestamptz,
    last_attempt_at timestamptz NOT NULL,
    revision bigint NOT NULL,
    retain_until timestamptz NOT NULL,
    CONSTRAINT iam_auth_throttle_buckets_pkey PRIMARY KEY (
        bucket_scope_kind, scope_id, purpose, dimension, key_id, bucket_hmac
    ),
    CONSTRAINT iam_auth_throttle_buckets_environment_fk FOREIGN KEY (environment_id)
        REFERENCES platform.environment_identity (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_auth_throttle_buckets_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_auth_throttle_buckets_authority_fk FOREIGN KEY (
        authority_scope_kind, authority_scope_id, authority_domain
    ) REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_auth_throttle_buckets_authority_group_fk FOREIGN KEY (
        authority_scope_kind, authority_scope_id, authority_group_id
    ) REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_auth_throttle_buckets_authority_member_fk FOREIGN KEY (
        authority_scope_kind, authority_scope_id,
        authority_group_id, authority_domain
    ) REFERENCES platform.cutover_group_members (
        scope_kind, scope_id, group_id, domain
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_auth_throttle_buckets_scope_kind_check CHECK (
        bucket_scope_kind IN ('tenant', 'environment')
    ),
    CONSTRAINT iam_auth_throttle_buckets_purpose_check CHECK (
        purpose IN (
            'login', 'refresh', 'password_reset', 'email_verification',
            'portal_setup', 'invitation', 'reauth'
        )
    ),
    CONSTRAINT iam_auth_throttle_buckets_dimension_check CHECK (
        dimension IN (
            'tenant_slug', 'identifier', 'ip', 'identifier_ip', 'selector'
        )
    ),
    CONSTRAINT iam_auth_throttle_buckets_key_check CHECK (
        pg_catalog.octet_length(key_id) BETWEEN 1 AND 128 AND
        key_id ~ '^[a-z0-9][a-z0-9._-]{0,127}$' AND
        pg_catalog.octet_length(bucket_hmac) = 32
    ),
    CONSTRAINT iam_auth_throttle_buckets_tenant_shape_check CHECK (
        bucket_scope_kind <> 'tenant' OR (
            environment_id IS NULL AND tenant_id IS NOT NULL AND
            tenant_id = scope_id AND authority_scope_kind IS NOT NULL AND
            authority_scope_kind = 'tenant' AND authority_scope_id IS NOT NULL AND
            authority_scope_id = scope_id AND authority_group_id IS NOT NULL AND
            authority_group_id = 'iam_system_resources' AND
            authority_group_generation IS NOT NULL AND
            authority_group_generation > 0 AND authority_domain IS NOT NULL AND
            authority_domain = 'iam' AND cutover_generation IS NOT NULL AND
            cutover_generation > 0
        )
    ),
    CONSTRAINT iam_auth_throttle_buckets_environment_shape_check CHECK (
        bucket_scope_kind <> 'environment' OR (
            environment_id IS NOT NULL AND environment_id = scope_id AND
            tenant_id IS NULL AND
            authority_scope_kind IS NULL AND authority_scope_id IS NULL AND
            authority_group_id IS NULL AND authority_group_generation IS NULL AND
            authority_domain IS NULL AND cutover_generation IS NULL
        )
    ),
    CONSTRAINT iam_auth_throttle_buckets_policy_check CHECK (
        policy_version > 0 AND attempt_count >= 0 AND revision > 0
    ),
    CONSTRAINT iam_auth_throttle_buckets_window_check CHECK (
        window_start < window_end AND
        window_start <= last_attempt_at AND last_attempt_at < window_end AND
        (blocked_until IS NULL OR last_attempt_at <= blocked_until) AND
        window_end <= retain_until AND
        (blocked_until IS NULL OR blocked_until <= retain_until)
    )
);

CREATE INDEX iam_auth_throttle_buckets_retention_idx
    ON platform.iam_auth_throttle_buckets (retain_until);

CREATE INDEX iam_auth_throttle_buckets_blocked_idx
    ON platform.iam_auth_throttle_buckets (
        bucket_scope_kind, scope_id, blocked_until
    ) WHERE blocked_until IS NOT NULL;
