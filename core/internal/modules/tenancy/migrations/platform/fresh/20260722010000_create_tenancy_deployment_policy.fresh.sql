-- +goose Up
CREATE TABLE platform.tenancy_deployment_policy (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    environment_id uuid NOT NULL UNIQUE,
    deployment_mode text NOT NULL
        CHECK (deployment_mode IN ('onpremise', 'saas')),
    maximum_customer_tenants bigint,
    created_at timestamptz NOT NULL DEFAULT statement_timestamp(),
    CONSTRAINT tenancy_deployment_policy_environment_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.environment_identity (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT tenancy_deployment_policy_capacity_ck CHECK (
        (deployment_mode = 'onpremise'
            AND maximum_customer_tenants IS NOT NULL
            AND maximum_customer_tenants = 0)
        OR (deployment_mode = 'saas' AND maximum_customer_tenants IS NULL)
    )
);
