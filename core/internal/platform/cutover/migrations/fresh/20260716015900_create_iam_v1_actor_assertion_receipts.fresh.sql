-- +goose Up
CREATE TABLE platform.iam_v1_actor_assertion_receipts (
    tenant_id uuid NOT NULL,
    jti uuid NOT NULL,
    subject_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    target_domain text NOT NULL,
    audience text NOT NULL,
    key_id text NOT NULL,
    permission_decision_sha256 bytea NOT NULL,
    target_binding_sha256 bytea NOT NULL,
    iam_group_generation bigint NOT NULL,
    iam_generation bigint NOT NULL,
    target_generation bigint NOT NULL,
    token_version bigint NOT NULL,
    authorization_version bigint NOT NULL,
    authorization_clock bigint NOT NULL,
    issued_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz NOT NULL,
    accounted_at timestamptz,
    PRIMARY KEY (tenant_id, jti),
    CONSTRAINT iam_v1_actor_receipts_uuid_check CHECK (
        tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid AND
        jti <> '00000000-0000-0000-0000-000000000000'::uuid AND
        subject_id <> '00000000-0000-0000-0000-000000000000'::uuid AND
        permission_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT iam_v1_actor_receipts_domain_check CHECK (
        target_domain IN (
            'analytics','billing','catalog','documents',
            'notifications','work','workflow'
        )
    ),
    CONSTRAINT iam_v1_actor_receipts_audience_check CHECK (
        audience = 'urn:harpy:v1-domain:' || target_domain
    ),
    CONSTRAINT iam_v1_actor_receipts_key_check
        CHECK (key_id ~ '^[a-z][a-z0-9_-]{0,62}$'),
    CONSTRAINT iam_v1_actor_receipts_digest_check CHECK (
        pg_catalog.octet_length(permission_decision_sha256) = 32 AND
        pg_catalog.octet_length(target_binding_sha256) = 32
    ),
    CONSTRAINT iam_v1_actor_receipts_version_check CHECK (
        iam_group_generation > 0 AND iam_generation > 0 AND
        target_generation > 0 AND token_version > 0 AND
        authorization_version > 0 AND authorization_clock >= 0
    ),
    CONSTRAINT iam_v1_actor_receipts_lifetime_check CHECK (
        expires_at = issued_at + INTERVAL '30 seconds' AND
        consumed_at >= issued_at - INTERVAL '2 seconds' AND
        consumed_at < expires_at AND
        (accounted_at IS NULL OR accounted_at >= consumed_at)
    )
);

CREATE INDEX iam_v1_actor_receipts_expiry_idx
    ON platform.iam_v1_actor_assertion_receipts (expires_at, tenant_id);

INSERT INTO platform.contract_retirements (
    contract_key, owner_module, domain, state, remaining_scope_count,
    hit_count, revision, updated_at
)
SELECT 'iam.v1_actor_assertion.v1', 'iam', 'iam', 'active',
       pg_catalog.count(*), 0, 1, pg_catalog.clock_timestamp()
FROM platform.domain_cutovers AS authority
JOIN platform.tenants AS tenant
  ON tenant.id = authority.scope_id AND tenant.deleted_at IS NULL
WHERE authority.scope_kind = 'tenant'
  AND authority.domain IN (
      'analytics','billing','catalog','documents',
      'notifications','work','workflow'
  )
  AND authority.state IN ('v1','shadow') AND authority.writer = 'v1';
