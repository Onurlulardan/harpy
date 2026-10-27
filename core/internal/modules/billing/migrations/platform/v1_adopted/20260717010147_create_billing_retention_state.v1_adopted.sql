-- +goose Up
CREATE TABLE platform.billing_legal_holds (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    state text NOT NULL,
    placed_by uuid NOT NULL,
    place_operation_id uuid NOT NULL,
    place_idempotency_key text COLLATE "C" NOT NULL,
    place_request_sha256 bytea NOT NULL,
    reason text NOT NULL,
    ticket_reference text COLLATE "C" NOT NULL,
    placed_at timestamptz NOT NULL,
    released_by uuid,
    release_operation_id uuid,
    release_idempotency_key text COLLATE "C",
    release_request_sha256 bytea,
    release_reason text,
    released_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT billing_legal_holds_pkey PRIMARY KEY (id),
    CONSTRAINT billing_legal_holds_tenant_id_key UNIQUE (tenant_id, id),
    CONSTRAINT billing_legal_holds_place_operation_key
        UNIQUE (tenant_id, place_operation_id),
    CONSTRAINT billing_legal_holds_place_idempotency_key
        UNIQUE (tenant_id, place_idempotency_key),
    CONSTRAINT billing_legal_holds_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_legal_holds_state_check CHECK (
        (state = 'active' AND revision = 1
            AND released_by IS NULL AND release_operation_id IS NULL
            AND release_idempotency_key IS NULL
            AND release_request_sha256 IS NULL AND release_reason IS NULL
            AND released_at IS NULL)
        OR (state = 'released' AND revision = 2
            AND released_by IS NOT NULL AND release_operation_id IS NOT NULL
            AND release_idempotency_key IS NOT NULL
            AND pg_catalog.octet_length(release_request_sha256) = 32
            AND pg_catalog.octet_length(release_reason) BETWEEN 1 AND 2048
            AND released_at >= placed_at)
    ),
    CONSTRAINT billing_legal_holds_shape_check CHECK (
        pg_catalog.octet_length(place_idempotency_key) BETWEEN 1 AND 128
        AND pg_catalog.octet_length(place_request_sha256) = 32
        AND pg_catalog.octet_length(reason) BETWEEN 1 AND 2048
        AND pg_catalog.octet_length(ticket_reference) BETWEEN 1 AND 256
        AND (release_idempotency_key IS NULL OR
            pg_catalog.octet_length(release_idempotency_key) BETWEEN 1 AND 128)
    )
);

CREATE UNIQUE INDEX billing_legal_holds_active_tenant_idx
    ON platform.billing_legal_holds (tenant_id) WHERE state = 'active';
CREATE UNIQUE INDEX billing_legal_holds_release_operation_idx
    ON platform.billing_legal_holds (tenant_id, release_operation_id)
    WHERE release_operation_id IS NOT NULL;
CREATE UNIQUE INDEX billing_legal_holds_release_idempotency_idx
    ON platform.billing_legal_holds (tenant_id, release_idempotency_key)
    WHERE release_idempotency_key IS NOT NULL;

CREATE TABLE platform.billing_retention_contexts (
    backend_pid integer NOT NULL,
    transaction_id xid8 NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    job_id uuid NOT NULL,
    opened_at timestamptz NOT NULL,
    CONSTRAINT billing_retention_contexts_pkey
        PRIMARY KEY (backend_pid, transaction_id),
    CONSTRAINT billing_retention_contexts_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_retention_contexts_pid_check CHECK (backend_pid > 0)
);

CREATE TABLE platform.billing_retention_checkpoints (
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    revision bigint NOT NULL,
    cycle_count bigint NOT NULL,
    last_job_id uuid NOT NULL,
    last_attempt_id uuid NOT NULL,
    raw_cutoff timestamptz NOT NULL,
    entitlement_cutoff timestamptz NOT NULL,
    financial_cutoff timestamptz NOT NULL,
    last_completed_at timestamptz NOT NULL,
    CONSTRAINT billing_retention_checkpoints_pkey PRIMARY KEY (tenant_id),
    CONSTRAINT billing_retention_checkpoints_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_retention_checkpoints_revision_check CHECK (
        revision > 0 AND cycle_count > 0
        AND financial_cutoff < entitlement_cutoff
        AND entitlement_cutoff < raw_cutoff
    )
);

CREATE TABLE platform.billing_retention_receipts (
    job_id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    attempt_id uuid NOT NULL,
    payload_reference text COLLATE "C" NOT NULL,
    payload_sha256 bytea NOT NULL,
    successor_job_id uuid NOT NULL,
    successor_run_at timestamptz NOT NULL,
    successor_dedupe_key text COLLATE "C" NOT NULL,
    raw_bodies_redacted bigint NOT NULL,
    raw_records_purged bigint NOT NULL,
    entitlement_purged bigint NOT NULL,
    financial_purged bigint NOT NULL,
    legal_hold_blocked bigint NOT NULL,
    checkpoint_revision bigint NOT NULL,
    receipt_sha256 bytea NOT NULL,
    completed_at timestamptz NOT NULL,
    CONSTRAINT billing_retention_receipts_pkey PRIMARY KEY (job_id),
    CONSTRAINT billing_retention_receipts_successor_key UNIQUE (successor_job_id),
    CONSTRAINT billing_retention_receipts_dedupe_key
        UNIQUE (tenant_id, successor_dedupe_key),
    CONSTRAINT billing_retention_receipts_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_retention_receipts_shape_check CHECK (
        pg_catalog.octet_length(payload_reference) BETWEEN 1 AND 2048
        AND pg_catalog.octet_length(payload_sha256) = 32
        AND pg_catalog.octet_length(successor_dedupe_key) BETWEEN 1 AND 256
        AND raw_bodies_redacted >= 0 AND raw_records_purged >= 0
        AND entitlement_purged >= 0 AND financial_purged >= 0
        AND legal_hold_blocked >= 0 AND checkpoint_revision > 0
        AND pg_catalog.octet_length(receipt_sha256) = 32
    )
);
