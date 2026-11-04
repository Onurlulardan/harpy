-- +goose Up
CREATE TABLE platform.tenancy_deployment_policy (
    environment_id uuid PRIMARY KEY CHECK (
        environment_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'::uuid
    ),
    deployment_mode text NOT NULL CHECK (deployment_mode IN ('onpremise', 'saas')),
    maximum_customer_tenants bigint,
    created_at timestamptz NOT NULL,
    CHECK (
        (deployment_mode = 'onpremise' AND maximum_customer_tenants = 0)
        OR (deployment_mode = 'saas' AND maximum_customer_tenants IS NULL)
    )
);

REVOKE ALL ON TABLE platform.tenancy_deployment_policy FROM PUBLIC;

GRANT SELECT, INSERT ON TABLE platform.tenancy_deployment_policy TO harpy_api, harpy_worker;

-- +goose Down
DROP TABLE platform.tenancy_deployment_policy;
