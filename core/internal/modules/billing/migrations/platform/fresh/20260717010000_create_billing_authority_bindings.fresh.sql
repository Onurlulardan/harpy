-- +goose Up
CREATE TABLE platform.billing_global_authorities (
    authority_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_group_id text NOT NULL,
    authority_group_generation bigint NOT NULL,
    authority_domain text NOT NULL,
    authority_domain_generation bigint NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_global_authorities_pkey PRIMARY KEY (authority_id),
    CONSTRAINT billing_global_authorities_tuple_key UNIQUE (
        scope_id, authority_group_generation, authority_domain_generation
    ),
    CONSTRAINT billing_global_authorities_environment_fk FOREIGN KEY (scope_id)
        REFERENCES platform.environment_identity (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_global_authorities_group_fk FOREIGN KEY (
        scope_kind, scope_id, authority_group_id
    ) REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_global_authorities_member_fk FOREIGN KEY (
        scope_kind, scope_id, authority_group_id, authority_domain
    ) REFERENCES platform.cutover_group_members (
        scope_kind, scope_id, group_id, domain
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_global_authorities_domain_fk FOREIGN KEY (
        scope_kind, scope_id, authority_domain
    ) REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_global_authorities_shape_check CHECK (
        authority_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND scope_kind = 'global'
        AND authority_group_id = 'billing'
        AND authority_domain = 'billing'
        AND authority_group_generation > 0
        AND authority_domain_generation > 0
    )
);

CREATE TABLE platform.billing_tenant_authorities (
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_group_id text NOT NULL,
    authority_group_generation bigint NOT NULL,
    billing_domain text NOT NULL,
    billing_generation bigint NOT NULL,
    forms_domain text NOT NULL,
    forms_generation bigint NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_tenant_authorities_pkey PRIMARY KEY (authority_id),
    CONSTRAINT billing_tenant_authorities_identity_key
        UNIQUE (authority_id, tenant_id),
    CONSTRAINT billing_tenant_authorities_tuple_key UNIQUE (
        tenant_id, authority_group_generation,
        billing_generation, forms_generation
    ),
    CONSTRAINT billing_tenant_authorities_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_tenant_authorities_group_fk FOREIGN KEY (
        scope_kind, scope_id, authority_group_id
    ) REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_tenant_authorities_billing_member_fk FOREIGN KEY (
        scope_kind, scope_id, authority_group_id, billing_domain
    ) REFERENCES platform.cutover_group_members (
        scope_kind, scope_id, group_id, domain
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_tenant_authorities_forms_member_fk FOREIGN KEY (
        scope_kind, scope_id, authority_group_id, forms_domain
    ) REFERENCES platform.cutover_group_members (
        scope_kind, scope_id, group_id, domain
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_tenant_authorities_billing_domain_fk FOREIGN KEY (
        scope_kind, scope_id, billing_domain
    ) REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_tenant_authorities_forms_domain_fk FOREIGN KEY (
        scope_kind, scope_id, forms_domain
    ) REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_tenant_authorities_shape_check CHECK (
        authority_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND scope_kind = 'tenant' AND scope_id = tenant_id
        AND authority_group_id = 'billing_forms'
        AND billing_domain = 'billing' AND forms_domain = 'forms'
        AND authority_group_generation > 0
        AND billing_generation > 0 AND forms_generation > 0
    )
);

