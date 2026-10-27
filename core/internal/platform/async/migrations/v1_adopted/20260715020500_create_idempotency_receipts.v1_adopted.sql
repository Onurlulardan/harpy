-- +goose Up
CREATE TABLE platform.idempotency_receipts (
    receipt_id uuid PRIMARY KEY,
    record_kind text NOT NULL CHECK (record_kind IN ('http', 'effect')),
    tenant_id uuid,
    scope_kind text NOT NULL CHECK (scope_kind IN ('global', 'tenant')),
    scope_id uuid NOT NULL,
    authority_group_id text NOT NULL
        CHECK (authority_group_id ~ '^[a-z][a-z0-9_]{0,62}$'),
    authority_group_generation bigint NOT NULL CHECK (authority_group_generation > 0),
    authority_domain text NOT NULL
        CHECK (authority_domain ~ '^[a-z][a-z0-9_]{0,62}$'),
    cutover_generation bigint NOT NULL CHECK (cutover_generation > 0),
    writer_origin text NOT NULL CHECK (writer_origin IN ('v1', 'v2')),
    request_sha256 text NOT NULL CHECK (request_sha256 ~ '^[0-9a-f]{64}$'),
    retention_class text NOT NULL CHECK (retention_class IN (
        'interactive_write', 'administrative_write', 'async_handle',
        'external_effect', 'workflow_effect'
    )),
    principal_id uuid,
    operation_id text CHECK (operation_id ~ '^[a-z][a-z0-9_.:-]{0,127}$'),
    idempotency_key text
        CHECK (pg_catalog.octet_length(idempotency_key) BETWEEN 1 AND 512),
    authorization_version bigint CHECK (authorization_version > 0),
    response_status integer CHECK (response_status BETWEEN 100 AND 599),
    response_headers jsonb,
    response_body bytea,
    http_result_ref text
        CHECK (pg_catalog.octet_length(http_result_ref) BETWEEN 1 AND 2048),
    response_sha256 text CHECK (response_sha256 ~ '^[0-9a-f]{64}$'),
    async_job_id uuid,
    effect_id uuid,
    effect_kind text CHECK (effect_kind ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    effect_revision bigint CHECK (effect_revision > 0),
    effect_key text CHECK (pg_catalog.octet_length(effect_key) BETWEEN 1 AND 512),
    provider text CHECK (provider ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    operation text CHECK (operation ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    correlation_id text CHECK (
        correlation_id ~ '^[A-Za-z0-9][A-Za-z0-9_.:-]{0,63}$'
    ),
    causation_id uuid,
    effect_status text CHECK (effect_status IN ('succeeded', 'failed')),
    effect_error_class text CHECK (effect_error_class IN (
        'transient', 'permanent', 'conflict', 'cancelled'
    )),
    effect_code text CHECK (effect_code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    provider_reference text
        CHECK (pg_catalog.octet_length(provider_reference) BETWEEN 1 AND 512),
    effect_result_ref text
        CHECK (pg_catalog.octet_length(effect_result_ref) BETWEEN 1 AND 2048),
    effect_result_sha256 text
        CHECK (effect_result_sha256 ~ '^[0-9a-f]{64}$'),
    completed_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    retain_until timestamptz NOT NULL,
    CONSTRAINT idempotency_receipts_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id) ON DELETE RESTRICT,
    CONSTRAINT idempotency_receipts_authority_fk
        FOREIGN KEY (scope_kind, scope_id, authority_domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON DELETE RESTRICT,
    CONSTRAINT idempotency_receipts_authority_group_fk
        FOREIGN KEY (scope_kind, scope_id, authority_group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON DELETE RESTRICT,
    CONSTRAINT idempotency_receipts_authority_member_fk
        FOREIGN KEY (scope_kind, scope_id, authority_group_id, authority_domain)
        REFERENCES platform.cutover_group_members (
            scope_kind, scope_id, group_id, domain
        ) ON DELETE RESTRICT,
    CONSTRAINT idempotency_receipts_async_job_fk
        FOREIGN KEY (scope_kind, scope_id, async_job_id)
        REFERENCES platform.jobs (scope_kind, scope_id, id) ON DELETE RESTRICT,
    CONSTRAINT idempotency_receipts_scope_check CHECK (
        (scope_kind = 'tenant' AND tenant_id IS NOT NULL AND tenant_id = scope_id) OR
        (scope_kind = 'global' AND tenant_id IS NULL)
    ),
    CONSTRAINT idempotency_receipts_http_shape_check CHECK (
        record_kind <> 'http' OR
        (retention_class IN (
            'interactive_write', 'administrative_write', 'async_handle'
        ) AND tenant_id IS NOT NULL AND principal_id IS NOT NULL AND
            operation_id IS NOT NULL AND idempotency_key IS NOT NULL AND
            authorization_version IS NOT NULL AND response_status IS NOT NULL AND
            response_headers IS NOT NULL AND
            pg_catalog.jsonb_typeof(response_headers) = 'object' AND
            pg_catalog.octet_length(response_headers::text) <= 16384 AND
            response_sha256 IS NOT NULL AND
            ((retention_class = 'async_handle' AND async_job_id IS NOT NULL) OR
             (retention_class <> 'async_handle' AND async_job_id IS NULL)) AND
            effect_id IS NULL AND
            effect_kind IS NULL AND effect_revision IS NULL AND
            effect_key IS NULL AND provider IS NULL AND
            operation IS NULL AND correlation_id IS NULL AND causation_id IS NULL AND
            effect_status IS NULL AND effect_error_class IS NULL AND
            effect_code IS NULL AND provider_reference IS NULL AND
            effect_result_ref IS NULL AND effect_result_sha256 IS NULL AND
            ((response_body IS NOT NULL AND http_result_ref IS NULL AND
                pg_catalog.octet_length(response_body) <= 65536) OR
             (response_body IS NULL AND http_result_ref IS NOT NULL)))
    ),
    CONSTRAINT idempotency_receipts_effect_shape_check CHECK (
        record_kind <> 'effect' OR
        (retention_class IN ('external_effect', 'workflow_effect') AND
            principal_id IS NULL AND operation_id IS NULL AND
            idempotency_key IS NULL AND authorization_version IS NULL AND
            response_status IS NULL AND response_headers IS NULL AND
            response_body IS NULL AND http_result_ref IS NULL AND
            response_sha256 IS NULL AND async_job_id IS NULL AND effect_id IS NOT NULL AND
            effect_kind IS NOT NULL AND effect_revision IS NOT NULL AND
            effect_key IS NOT NULL AND
            provider IS NOT NULL AND operation IS NOT NULL AND
            correlation_id IS NOT NULL AND causation_id IS NOT NULL AND
            effect_status IS NOT NULL AND effect_code IS NOT NULL AND
            ((effect_status = 'succeeded' AND effect_error_class IS NULL AND
                effect_result_ref IS NOT NULL AND effect_result_sha256 IS NOT NULL) OR
             (effect_status = 'failed' AND effect_error_class IS NOT NULL AND
                effect_result_ref IS NULL AND effect_result_sha256 IS NULL)))
    ),
    CONSTRAINT idempotency_receipts_time_check CHECK (completed_at < retain_until)
);

ALTER TABLE platform.job_attempts
    ADD CONSTRAINT job_attempts_effect_receipt_fk
    FOREIGN KEY (effect_receipt_id)
    REFERENCES platform.idempotency_receipts (receipt_id) ON DELETE RESTRICT;
