-- +goose Up
CREATE TABLE platform.tenancy_deployment_policy_initializations (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL UNIQUE,
    operation_kind text NOT NULL
        CHECK (operation_kind = 'environment_initialize'),
    deployment_mode text NOT NULL
        CHECK (deployment_mode IN ('onpremise', 'saas')),
    maximum_customer_tenants bigint,
    created_at timestamptz NOT NULL DEFAULT statement_timestamp(),
    CONSTRAINT tenancy_policy_initialization_policy_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.tenancy_deployment_policy (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT tenancy_policy_initialization_capacity_ck CHECK (
        (deployment_mode = 'onpremise'
            AND maximum_customer_tenants = 0)
        OR (deployment_mode = 'saas'
            AND maximum_customer_tenants IS NULL)
    )
);
