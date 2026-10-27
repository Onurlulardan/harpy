-- +goose Up
CREATE TABLE analytics_v1_adoption_stage (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    legacy_endpoint_id uuid NOT NULL,
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256)=32),
    source_document jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(source_document)='object'),
    disposition text NOT NULL CHECK (disposition IN ('adopted','blocked','retired')),
    reason_code text,
    target_document jsonb,
    target_sha256 bytea,
    removed boolean NOT NULL,
    source_xid xid8 NOT NULL,
    commit_lsn pg_lsn NOT NULL,
    batch_ordinal integer NOT NULL CHECK (batch_ordinal BETWEEN 1 AND 2147483),
    change_ordinal smallint NOT NULL CHECK (change_ordinal BETWEEN 1 AND 1000),
    recorded_at timestamptz NOT NULL,
    PRIMARY KEY(operation_id,run_id,legacy_endpoint_id),
    CHECK (pg_catalog.sha256(pg_catalog.convert_to(source_document::text,'UTF8'))=source_sha256),
    CHECK ((disposition='adopted' AND reason_code IS NULL
        AND pg_catalog.jsonb_typeof(target_document)='object'
        AND pg_catalog.octet_length(target_sha256)=32
        AND pg_catalog.sha256(pg_catalog.convert_to(target_document::text,'UTF8'))=target_sha256)
      OR (disposition IN ('blocked','retired')
        AND pg_catalog.octet_length(reason_code) BETWEEN 1 AND 64
        AND target_document IS NULL AND target_sha256 IS NULL)),
    CHECK (date_trunc('microseconds',recorded_at)=recorded_at)
);

CREATE TABLE analytics_v1_adoption_differentials (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    legacy_endpoint_id uuid NOT NULL,
    source_result_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_result_sha256)=32),
    target_result_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_result_sha256)=32),
    wire_contract_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(wire_contract_sha256)=32),
    compatible_image_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(compatible_image_sha256)=32),
    reverse_source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(reverse_source_sha256)=32),
    evidence_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(evidence_sha256)=32),
    recorded_at timestamptz NOT NULL,
    PRIMARY KEY(operation_id,run_id,legacy_endpoint_id),
    FOREIGN KEY(operation_id,run_id,legacy_endpoint_id)
      REFERENCES analytics_v1_adoption_stage(operation_id,run_id,legacy_endpoint_id)
      ON DELETE RESTRICT,
    CHECK (source_result_sha256=target_result_sha256),
    CHECK (date_trunc('microseconds',recorded_at)=recorded_at)
);

CREATE TABLE analytics_v1_adoption_receipts (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    cutover_generation bigint NOT NULL CHECK (cutover_generation>0),
    plan_sha256 text NOT NULL CHECK (plan_sha256~'^[0-9a-f]{64}$'),
    transform_sha256 text NOT NULL CHECK (transform_sha256~'^[0-9a-f]{64}$'),
    lease_manifest_sha256 text NOT NULL CHECK (lease_manifest_sha256~'^[0-9a-f]{64}$'),
    final_lsn pg_lsn NOT NULL,
    source_count bigint NOT NULL CHECK (source_count>=0),
    source_root_sha256 text NOT NULL CHECK (source_root_sha256~'^[0-9a-f]{64}$'),
    target_count bigint NOT NULL CHECK (target_count>=0),
    target_root_sha256 text NOT NULL CHECK (target_root_sha256~'^[0-9a-f]{64}$'),
    dependency_sha256 text NOT NULL CHECK (dependency_sha256~'^[0-9a-f]{64}$'),
    public_approval_sha256 text NOT NULL CHECK (public_approval_sha256~'^[0-9a-f]{64}$'),
    zero_hit_sha256 text NOT NULL CHECK (zero_hit_sha256~'^[0-9a-f]{64}$'),
    rollback_sha256 text NOT NULL CHECK (rollback_sha256~'^[0-9a-f]{64}$'),
    watermark_sha256 text NOT NULL CHECK (watermark_sha256~'^[0-9a-f]{64}$'),
    evidence_sha256 text NOT NULL CHECK (evidence_sha256~'^[0-9a-f]{64}$'),
    acceptance_sha256 text NOT NULL CHECK (acceptance_sha256~'^[0-9a-f]{64}$'),
    recorded_at timestamptz NOT NULL,
    PRIMARY KEY(operation_id,run_id),
    CHECK (date_trunc('microseconds',recorded_at)=recorded_at)
);
