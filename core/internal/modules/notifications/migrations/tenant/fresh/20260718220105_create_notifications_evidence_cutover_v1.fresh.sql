-- +goose Up
CREATE TABLE notifications_audit_entries (
    id uuid PRIMARY KEY,
    action text COLLATE "C" NOT NULL
        CHECK (action ~ '^notifications[.][a-z0-9_.]{2,95}$'),
    resource_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    resource_revision bigint NOT NULL CHECK (resource_revision > 0),
    occurred_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(occurred_at))
);

CREATE TABLE notifications_realtime_hints (
    id uuid PRIMARY KEY,
    message_id uuid NOT NULL,
    recipient_id uuid NOT NULL,
    inbox_revision bigint NOT NULL CHECK (inbox_revision > 0),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    published_at timestamptz,
    CONSTRAINT notifications_realtime_hints_message_key UNIQUE (message_id),
    CONSTRAINT notifications_realtime_hints_message_fk FOREIGN KEY (message_id)
        REFERENCES notifications_messages(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (published_at IS NULL OR pg_catalog.isfinite(published_at))
);

CREATE INDEX notifications_realtime_hints_pending_idx
    ON notifications_realtime_hints (created_at, id) WHERE published_at IS NULL;

CREATE TABLE notifications_operator_alerts (
    id uuid PRIMARY KEY,
    code text COLLATE "C" NOT NULL CHECK (code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    resource_id uuid NOT NULL,
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    acknowledged_at timestamptz
);

CREATE TABLE notifications_writer_authority (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    writer text COLLATE "C" NOT NULL CHECK (writer IN ('unassigned','v1','v2')),
    lifecycle_state text COLLATE "C" NOT NULL
        CHECK (lifecycle_state IN ('bootstrap','writable','quiescing','fenced')),
    generation bigint NOT NULL CHECK (generation >= 0),
    row_version bigint NOT NULL CHECK (row_version > 0),
    updated_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(updated_at)),
    CHECK ((writer = 'unassigned') = (lifecycle_state = 'bootstrap'))
);

INSERT INTO notifications_writer_authority (
    singleton, writer, lifecycle_state, generation, row_version, updated_at
) VALUES (true, 'unassigned', 'bootstrap', 0, 1, '2026-07-18T00:00:00Z');

CREATE TABLE notifications_adoption_runs (
    id uuid PRIMARY KEY,
    source_generation text COLLATE "C" NOT NULL
        CHECK (pg_catalog.octet_length(source_generation) BETWEEN 1 AND 128),
    source_manifest_digest bytea NOT NULL CHECK (pg_catalog.octet_length(source_manifest_digest) = 32),
    lifecycle_state text COLLATE "C" NOT NULL
        CHECK (lifecycle_state IN ('planned','applying','verified','failed','rolled_back')),
    started_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(started_at)),
    completed_at timestamptz,
    row_version bigint NOT NULL CHECK (row_version > 0),
    CONSTRAINT notifications_adoption_runs_source_key
        UNIQUE (source_generation, source_manifest_digest)
);

CREATE TABLE notifications_adoption_checkpoints (
    run_id uuid NOT NULL,
    source_table text COLLATE "C" NOT NULL
        CHECK (source_table IN ('notifications','email_flows')),
    last_source_id uuid,
    adopted_count bigint NOT NULL CHECK (adopted_count >= 0),
    source_checksum bytea NOT NULL CHECK (pg_catalog.octet_length(source_checksum) = 32),
    target_checksum bytea NOT NULL CHECK (pg_catalog.octet_length(target_checksum) = 32),
    updated_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(updated_at)),
    PRIMARY KEY (run_id, source_table),
    CONSTRAINT notifications_adoption_checkpoints_run_fk FOREIGN KEY (run_id)
        REFERENCES notifications_adoption_runs(id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE notifications_adoption_issues (
    id uuid PRIMARY KEY,
    run_id uuid NOT NULL,
    source_table text COLLATE "C" NOT NULL,
    source_key_digest bytea NOT NULL CHECK (pg_catalog.octet_length(source_key_digest) = 32),
    code text COLLATE "C" NOT NULL CHECK (code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    resolved_at timestamptz,
    CONSTRAINT notifications_adoption_issues_run_fk FOREIGN KEY (run_id)
        REFERENCES notifications_adoption_runs(id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE notifications_migration_intents (
    producer text COLLATE "C" NOT NULL
        CHECK (pg_catalog.octet_length(producer) BETWEEN 1 AND 64),
    effect_key_hash bytea NOT NULL CHECK (pg_catalog.octet_length(effect_key_hash) = 32),
    recipient_id uuid NOT NULL,
    producer_generation bigint NOT NULL CHECK (producer_generation > 0),
    notification_generation bigint NOT NULL CHECK (notification_generation > 0),
    canonical_digest bytea NOT NULL CHECK (pg_catalog.octet_length(canonical_digest) = 32),
    payload_digest bytea NOT NULL CHECK (pg_catalog.octet_length(payload_digest) = 32),
    payload jsonb NOT NULL,
    job_id uuid NOT NULL UNIQUE,
    state text COLLATE "C" NOT NULL
        CHECK (state IN ('queued','accepted','failed','outcome_unknown')),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    updated_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(updated_at)),
    row_version bigint NOT NULL CHECK (row_version > 0),
    PRIMARY KEY (producer, effect_key_hash, recipient_id),
    CHECK (pg_catalog.jsonb_typeof(payload) = 'object'),
    CHECK (pg_catalog.pg_column_size(payload) <= 65536),
    CHECK (updated_at >= created_at)
);

CREATE TABLE notifications_migration_receipts (
    producer text COLLATE "C" NOT NULL,
    effect_key_hash bytea NOT NULL CHECK (pg_catalog.octet_length(effect_key_hash) = 32),
    recipient_id uuid NOT NULL,
    producer_generation bigint NOT NULL CHECK (producer_generation > 0),
    notification_generation bigint NOT NULL CHECK (notification_generation > 0),
    v1_receipt_digest bytea NOT NULL CHECK (pg_catalog.octet_length(v1_receipt_digest) = 32),
    terminal_outcome text COLLATE "C" NOT NULL
        CHECK (terminal_outcome IN ('accepted','failed','outcome_unknown')),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    PRIMARY KEY (producer, effect_key_hash, recipient_id),
    CONSTRAINT notifications_migration_receipts_intent_fk FOREIGN KEY (
        producer, effect_key_hash, recipient_id
    ) REFERENCES notifications_migration_intents (
        producer, effect_key_hash, recipient_id
    ) ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE notifications_compatibility_hits (
    id uuid PRIMARY KEY,
    contract_name text COLLATE "C" NOT NULL
        CHECK (contract_name IN ('notifications_actor_stream_v1','notification_intent_bridge_v1')),
    consumer_manifest_digest bytea NOT NULL
        CHECK (pg_catalog.octet_length(consumer_manifest_digest) = 32),
    occurred_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(occurred_at))
);

CREATE TABLE notifications_cutover_evidence (
    id uuid PRIMARY KEY,
    run_id uuid NOT NULL,
    source_queue_depth bigint NOT NULL CHECK (source_queue_depth >= 0),
    future_delivery_count bigint NOT NULL CHECK (future_delivery_count >= 0),
    running_delivery_count bigint NOT NULL CHECK (running_delivery_count >= 0),
    unresolved_provider_receipt_count bigint NOT NULL
        CHECK (unresolved_provider_receipt_count >= 0),
    legacy_connection_count bigint NOT NULL CHECK (legacy_connection_count >= 0),
    source_evidence_digest bytea NOT NULL
        CHECK (pg_catalog.octet_length(source_evidence_digest) = 32),
    observed_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(observed_at)),
    expires_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(expires_at)),
    CONSTRAINT notifications_cutover_evidence_run_fk FOREIGN KEY (run_id)
        REFERENCES notifications_adoption_runs(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT notifications_cutover_evidence_identity_key
        UNIQUE (run_id, source_evidence_digest),
    CHECK (expires_at = observed_at + interval '5 minutes')
);
