-- +goose Up
CREATE TABLE platform.analytics_tenant_authorities (
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_group_id text NOT NULL,
    authority_group_generation bigint NOT NULL,
    authority_domain text NOT NULL,
    authority_domain_generation bigint NOT NULL,
    authority_mode text NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT analytics_tenant_authorities_pkey PRIMARY KEY (authority_id),
    CONSTRAINT analytics_tenant_authorities_identity_key UNIQUE (authority_id, tenant_id),
    CONSTRAINT analytics_tenant_authorities_generation_key UNIQUE (
        tenant_id, authority_group_generation, authority_domain_generation
    ),
    CONSTRAINT analytics_tenant_authorities_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_tenant_authorities_group_fk FOREIGN KEY (
        scope_kind, scope_id, authority_group_id
    ) REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_tenant_authorities_member_fk FOREIGN KEY (
        scope_kind, scope_id, authority_group_id, authority_domain
    ) REFERENCES platform.cutover_group_members (
        scope_kind, scope_id, group_id, domain
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_tenant_authorities_domain_fk FOREIGN KEY (
        scope_kind, scope_id, authority_domain
    ) REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_tenant_authorities_shape_check CHECK (
        pg_catalog.uuid_extract_version(authority_id) = 7
        AND scope_kind = 'tenant' AND scope_id = tenant_id
        AND authority_group_id = 'analytics' AND authority_domain = 'analytics'
        AND authority_group_generation > 0 AND authority_domain_generation > 0
        AND authority_mode IN ('runtime_v2', 'adopter_shadow')
        AND pg_catalog.isfinite(created_at)
    )
);

