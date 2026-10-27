-- +goose Up
CREATE TABLE documents_v1_adoption_stage (
    operation_id uuid NOT NULL,
    cutover_run_id uuid NOT NULL,
    source_kind text NOT NULL CHECK (source_kind IN (
        'folder','item','version','inventory'
    )),
    source_identity text NOT NULL CHECK (
        pg_catalog.octet_length(source_identity) BETWEEN 1 AND 1024
        AND source_identity OPERATOR(pg_catalog.!~) '[[:cntrl:]]'
    ),
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256)=32),
    target_kind text NOT NULL CHECK (target_kind IN (
        'folder','item','version','inventory','checkpoint'
    )),
    target_identity text NOT NULL CHECK (
        pg_catalog.octet_length(target_identity) BETWEEN 1 AND 1024
        AND target_identity OPERATOR(pg_catalog.!~) '[[:cntrl:]]'
    ),
    target_id uuid NOT NULL CHECK (
        target_id<>'00000000-0000-0000-0000-000000000000'::uuid
    ),
    target_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_sha256)=32),
    target_record jsonb NOT NULL CHECK (
        pg_catalog.jsonb_typeof(target_record)='object' AND target_record<>'{}'::jsonb
    ),
    removed boolean NOT NULL,
    source_xid xid8 NOT NULL,
    commit_lsn pg_lsn NOT NULL,
    batch_ordinal integer NOT NULL CHECK (batch_ordinal BETWEEN 1 AND 2147483),
    change_ordinal integer NOT NULL CHECK (change_ordinal BETWEEN 1 AND 1000),
    recorded_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(recorded_at)),
    PRIMARY KEY (operation_id,cutover_run_id,target_kind,target_identity),
    CONSTRAINT documents_v1_stage_target_digest_check CHECK (
        target_sha256=pg_catalog.sha256(
            pg_catalog.convert_to(target_record::text,'UTF8')
        )
    )
);

CREATE INDEX documents_v1_adoption_stage_source_idx
ON documents_v1_adoption_stage (
    operation_id,cutover_run_id,source_kind,source_identity,target_kind
);

CREATE TABLE documents_v1_cutover_receipts (
    operation_id uuid NOT NULL,
    cutover_run_id uuid NOT NULL,
    adoption_run_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    cutover_generation bigint NOT NULL CHECK (cutover_generation>0),
    source_count bigint NOT NULL CHECK (source_count>=0),
    target_count bigint NOT NULL CHECK (target_count>=0),
    source_plaintext_bytes bigint NOT NULL CHECK (source_plaintext_bytes>=0),
    target_plaintext_bytes bigint NOT NULL CHECK (target_plaintext_bytes>=0),
    source_merkle_sha256 text NOT NULL CHECK (source_merkle_sha256~'^[0-9a-f]{64}$'),
    target_merkle_sha256 text NOT NULL CHECK (target_merkle_sha256~'^[0-9a-f]{64}$'),
    row_root_sha256 text NOT NULL CHECK (row_root_sha256~'^[0-9a-f]{64}$'),
    binding_root_sha256 text NOT NULL CHECK (binding_root_sha256~'^[0-9a-f]{64}$'),
    effect_root_sha256 text NOT NULL CHECK (effect_root_sha256~'^[0-9a-f]{64}$'),
    final_lsn pg_lsn NOT NULL,
    accepted_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(accepted_at)),
    receipt_sha256 text NOT NULL CHECK (receipt_sha256~'^[0-9a-f]{64}$'),
    PRIMARY KEY (operation_id,cutover_run_id),
    UNIQUE (receipt_sha256)
);

CREATE TABLE documents_v1_reverse_receipts (
    reconciliation_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    source_generation bigint NOT NULL CHECK (source_generation>0),
    target_generation bigint NOT NULL CHECK (target_generation>0),
    source_count bigint NOT NULL CHECK (source_count>=0),
    target_count bigint NOT NULL CHECK (target_count>=0),
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256)=32),
    target_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_sha256)=32),
    requested_by uuid NOT NULL,
    approved_by uuid NOT NULL CHECK (approved_by<>requested_by),
    accepted_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(accepted_at)),
    receipt_sha256 text NOT NULL CHECK (receipt_sha256~'^[0-9a-f]{64}$'),
    UNIQUE (receipt_sha256),
    CHECK (source_count=target_count AND source_sha256=target_sha256)
);
