-- +goose Up
CREATE TABLE platform.iam_refresh_locators (
    selector bytea PRIMARY KEY,
    tenant_id uuid NOT NULL,
    token_id uuid NOT NULL,
    family_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_group_id text NOT NULL,
    authority_group_generation bigint NOT NULL,
    authority_domain text NOT NULL,
    cutover_generation bigint NOT NULL,
    source_revision bigint NOT NULL,
    source_sha256 bytea NOT NULL,
    source_token_state text NOT NULL,
    issued_at timestamptz NOT NULL,
    projected_at timestamptz NOT NULL,
    family_absolute_expires_at timestamptz NOT NULL,
    retain_until timestamptz NOT NULL,
    CONSTRAINT iam_refresh_locators_token_unique UNIQUE (tenant_id, token_id),
    CONSTRAINT iam_refresh_locators_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_refresh_locators_authority_fk FOREIGN KEY (
        scope_kind, scope_id, authority_domain
    ) REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_refresh_locators_authority_group_fk FOREIGN KEY (
        scope_kind, scope_id, authority_group_id
    ) REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_refresh_locators_authority_member_fk FOREIGN KEY (
        scope_kind, scope_id, authority_group_id, authority_domain
    ) REFERENCES platform.cutover_group_members (
        scope_kind, scope_id, group_id, domain
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_refresh_locators_selector_check CHECK (
        pg_catalog.octet_length(selector) = 16
    ),
    CONSTRAINT iam_refresh_locators_scope_check CHECK (
        scope_kind = 'tenant' AND scope_id = tenant_id AND
        authority_group_id = 'iam_system_resources' AND
        authority_domain = 'iam'
    ),
    CONSTRAINT iam_refresh_locators_generation_check CHECK (
        authority_group_generation > 0 AND cutover_generation > 0
    ),
    CONSTRAINT iam_refresh_locators_source_check CHECK (
        source_revision > 0 AND pg_catalog.octet_length(source_sha256) = 32
    ),
    CONSTRAINT iam_refresh_locators_source_state_check CHECK (
        source_token_state IN ('active', 'consumed', 'revoked', 'expired')
    ),
    CONSTRAINT iam_refresh_locators_time_check CHECK (
        issued_at <= projected_at AND issued_at < family_absolute_expires_at AND
        projected_at < retain_until AND
        retain_until >= family_absolute_expires_at + INTERVAL '24 hours'
    )
);

CREATE INDEX iam_refresh_locators_family_idx
    ON platform.iam_refresh_locators (tenant_id, family_id, retain_until);

CREATE INDEX iam_refresh_locators_retention_idx
    ON platform.iam_refresh_locators (retain_until);
