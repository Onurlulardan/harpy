-- +goose Up
CREATE TABLE notifications_v1_adoption_stage (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    source_kind text NOT NULL CHECK (source_kind = 'notification'),
    source_identity text NOT NULL CHECK (
        pg_catalog.octet_length(source_identity) BETWEEN 1 AND 512
        AND source_identity OPERATOR(pg_catalog.!~) '[[:cntrl:]]'
    ),
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256) = 32),
    target_kind text NOT NULL CHECK (target_kind IN (
        'message','effect','delivery','attempt','receipt'
    )),
    target_identity text NOT NULL CHECK (
        pg_catalog.octet_length(target_identity) BETWEEN 1 AND 1024
        AND target_identity OPERATOR(pg_catalog.!~) '[[:cntrl:]]'
    ),
    target_id uuid NOT NULL CHECK (
        target_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    target_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_sha256) = 32),
    target_record jsonb NOT NULL CHECK (
        pg_catalog.jsonb_typeof(target_record) = 'object' AND target_record <> '{}'::jsonb
    ),
    removed boolean NOT NULL,
    source_xid xid8 NOT NULL,
    commit_lsn pg_lsn NOT NULL,
    batch_ordinal integer NOT NULL CHECK (batch_ordinal BETWEEN 1 AND 2147483),
    change_ordinal integer NOT NULL CHECK (change_ordinal BETWEEN 1 AND 1000),
    recorded_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(recorded_at)),
    PRIMARY KEY (operation_id,run_id,target_kind,target_identity),
    CONSTRAINT notifications_v1_adoption_stage_digest_check CHECK (
        target_sha256 = pg_catalog.sha256(
            pg_catalog.convert_to(target_record::text,'UTF8')
        )
    )
);

CREATE INDEX notifications_v1_adoption_stage_source_idx
ON notifications_v1_adoption_stage (
    operation_id,run_id,source_identity,target_kind,target_identity
);

CREATE TABLE notifications_v1_cutover_receipts (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    cutover_generation bigint NOT NULL CHECK (cutover_generation > 0),
    source_count bigint NOT NULL CHECK (source_count >= 0),
    target_count bigint NOT NULL CHECK (target_count >= source_count),
    source_root_sha256 text NOT NULL CHECK (source_root_sha256 ~ '^[0-9a-f]{64}$'),
    target_root_sha256 text NOT NULL CHECK (target_root_sha256 ~ '^[0-9a-f]{64}$'),
    queue_evidence_sha256 text NOT NULL CHECK (queue_evidence_sha256 ~ '^[0-9a-f]{64}$'),
    provider_evidence_sha256 text NOT NULL
        CHECK (provider_evidence_sha256 ~ '^[0-9a-f]{64}$'),
    legacy_evidence_sha256 text NOT NULL CHECK (legacy_evidence_sha256 ~ '^[0-9a-f]{64}$'),
    bridge_state text NOT NULL CHECK (
        bridge_state = 'disabled_missing_signed_mtls_contract'
    ),
    final_lsn pg_lsn NOT NULL,
    projected_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(projected_at)),
    receipt_sha256 text NOT NULL CHECK (receipt_sha256 ~ '^[0-9a-f]{64}$'),
    PRIMARY KEY (operation_id,run_id),
    CONSTRAINT notifications_v1_cutover_receipts_digest_unique UNIQUE (receipt_sha256)
);
