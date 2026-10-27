-- +goose Up
CREATE TABLE platform.billing_catalog_operations (
    operation_id uuid NOT NULL,
    authority_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    idempotency_key text COLLATE "C" NOT NULL,
    request_sha256 bytea NOT NULL,
    aggregate_kind text NOT NULL,
    aggregate_id uuid NOT NULL,
    operation_kind text NOT NULL,
    result_version_id uuid,
    result_revision bigint NOT NULL,
    result_version bigint,
    result_lifecycle text NOT NULL,
    reason text,
    occurred_at timestamptz NOT NULL,
    CONSTRAINT billing_catalog_operations_pkey PRIMARY KEY (operation_id),
    CONSTRAINT billing_catalog_operations_idempotency_key UNIQUE (
        authority_id, actor_id, idempotency_key
    ),
    CONSTRAINT billing_catalog_operations_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_catalog_operations_shape_check CHECK (
        operation_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND actor_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND aggregate_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND pg_catalog.octet_length(idempotency_key) BETWEEN 1 AND 128
        AND pg_catalog.btrim(idempotency_key) = idempotency_key
        AND pg_catalog.octet_length(request_sha256) = 32
        AND aggregate_kind IN ('offering', 'discount', 'campaign')
        AND operation_kind IN ('create', 'create_version', 'retire')
        AND result_revision > 0
        AND result_lifecycle IN ('active', 'retired')
    ),
    CONSTRAINT billing_catalog_operations_result_check CHECK (
        (operation_kind IN ('create', 'create_version')
            AND result_version_id IS NOT NULL AND result_version > 0
            AND result_lifecycle = 'active' AND reason IS NULL)
        OR (operation_kind = 'retire'
            AND result_version_id IS NULL AND result_version IS NULL
            AND result_lifecycle = 'retired'
            AND pg_catalog.octet_length(reason) BETWEEN 1 AND 512)
    )
);

