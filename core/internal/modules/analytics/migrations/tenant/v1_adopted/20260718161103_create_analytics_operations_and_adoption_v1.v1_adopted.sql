-- +goose Up
CREATE TABLE analytics_operations (
    operation_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    operation_kind text NOT NULL,
    idempotency_key text,
    request_sha256 bytea NOT NULL,
    resource_kind text NOT NULL,
    resource_id uuid NOT NULL,
    result_revision bigint NOT NULL,
    result_sha256 bytea NOT NULL,
    occurred_at timestamptz NOT NULL,
    CONSTRAINT analytics_operations_pkey PRIMARY KEY (operation_id),
    CONSTRAINT analytics_operations_idempotency_key UNIQUE (
        actor_id, operation_kind, idempotency_key
    ),
    CONSTRAINT analytics_operations_shape_check CHECK (
        pg_catalog.uuid_extract_version(operation_id) = 7
        AND actor_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND operation_kind IN (
            'query_create', 'query_update', 'query_delete',
            'query_publish', 'query_revalidate',
            'dashboard_create', 'dashboard_update', 'dashboard_delete',
            'dashboard_publish', 'dashboard_grant_acl',
            'dashboard_revoke_acl', 'dashboard_set_default',
            'publication_create', 'publication_submit',
            'publication_approve', 'publication_activate',
            'publication_revoke'
        )
        AND (idempotency_key IS NULL OR (
            pg_catalog.octet_length(idempotency_key) BETWEEN 1 AND 128
            AND idempotency_key ~ '^[A-Za-z0-9_.:-]+$'
        ))
        AND pg_catalog.octet_length(request_sha256) = 32
        AND resource_kind IN ('query', 'dashboard', 'publication')
        AND resource_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND result_revision > 0
        AND pg_catalog.octet_length(result_sha256) = 32
        AND pg_catalog.isfinite(occurred_at)
    )
);

CREATE TABLE analytics_v1_adoption_items (
    run_id uuid NOT NULL,
    legacy_endpoint_id uuid NOT NULL,
    source_sha256 bytea NOT NULL,
    disposition text NOT NULL,
    reason_code text,
    query_id uuid,
    query_version_id uuid,
    publication_release_id uuid,
    target_sha256 bytea,
    created_at timestamptz NOT NULL,
    CONSTRAINT analytics_v1_adoption_items_pkey PRIMARY KEY (run_id, legacy_endpoint_id),
    CONSTRAINT analytics_v1_adoption_items_shape_check CHECK (
        run_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND legacy_endpoint_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND pg_catalog.octet_length(source_sha256) = 32
        AND disposition IN ('adopted', 'blocked', 'retired')
        AND ((disposition = 'adopted'
              AND pg_catalog.uuid_extract_version(query_id) = 7
              AND pg_catalog.uuid_extract_version(query_version_id) = 7
              AND pg_catalog.octet_length(target_sha256) = 32
              AND reason_code IS NULL)
          OR (disposition IN ('blocked', 'retired')
              AND pg_catalog.octet_length(reason_code) BETWEEN 1 AND 64
              AND query_id IS NULL AND query_version_id IS NULL
              AND publication_release_id IS NULL AND target_sha256 IS NULL))
        AND (publication_release_id IS NULL
             OR pg_catalog.uuid_extract_version(publication_release_id) = 7)
        AND pg_catalog.isfinite(created_at)
    )
);

CREATE TABLE analytics_v1_adoption_acceptance_receipts (
    run_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    cutover_generation bigint NOT NULL,
    source_count bigint NOT NULL,
    adopted_count bigint NOT NULL,
    retired_count bigint NOT NULL,
    root_sha256 bytea NOT NULL,
    accepted_at timestamptz NOT NULL,
    CONSTRAINT analytics_v1_adoption_acceptance_receipts_pkey PRIMARY KEY (run_id),
    CONSTRAINT analytics_v1_adoption_acceptance_shape_check CHECK (
        tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND cutover_generation > 0
        AND source_count >= 0 AND adopted_count >= 0 AND retired_count >= 0
        AND source_count = adopted_count + retired_count
        AND pg_catalog.octet_length(root_sha256) = 32
        AND pg_catalog.isfinite(accepted_at)
    )
);

