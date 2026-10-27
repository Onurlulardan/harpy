-- +goose Up
CREATE TABLE platform.iam_security_epochs (
    environment_id uuid PRIMARY KEY,
    security_epoch bytea NOT NULL
        CHECK (pg_catalog.octet_length(security_epoch) = 32),
    generation bigint NOT NULL CHECK (generation > 0),
    created_at timestamptz NOT NULL,
    rotated_at timestamptz NOT NULL CHECK (rotated_at >= created_at),
    CONSTRAINT iam_security_epochs_environment_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.tenancy_deployment_policy (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

INSERT INTO platform.iam_security_epochs (
    environment_id, security_epoch, generation, created_at, rotated_at
)
SELECT policy.environment_id,
       pg_catalog.sha256(
           pg_catalog.uuid_send(pg_catalog.gen_random_uuid()) ||
           pg_catalog.uuid_send(pg_catalog.gen_random_uuid())
       ),
       1, pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp()
FROM platform.tenancy_deployment_policy AS policy;
