-- +goose Up
CREATE TABLE platform.tenancy_tenant_bindings (
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    purpose text NOT NULL CHECK (purpose IN (
        'onpremise_primary', 'saas_control', 'saas_customer'
    )),
    created_at timestamptz NOT NULL DEFAULT statement_timestamp(),
    CONSTRAINT tenancy_tenant_bindings_pk
        PRIMARY KEY (environment_id, tenant_id),
    CONSTRAINT tenancy_tenant_bindings_tenant_uq UNIQUE (tenant_id),
    CONSTRAINT tenancy_tenant_bindings_policy_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.tenancy_deployment_policy (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT tenancy_tenant_bindings_tenant_fk
        FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE UNIQUE INDEX tenancy_tenant_bindings_primary_uq
ON platform.tenancy_tenant_bindings (environment_id)
WHERE purpose IN ('onpremise_primary', 'saas_control');
