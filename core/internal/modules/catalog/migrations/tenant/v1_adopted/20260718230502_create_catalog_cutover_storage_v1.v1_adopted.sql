-- +goose Up
CREATE TABLE catalog_v1_adoption_stage (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    source_kind text NOT NULL CHECK (source_kind IN (
        'modules','module_dependencies','module_forms',
        'module_schema_versions','module_migration_runs','menu_configs'
    )),
    source_identity text NOT NULL CHECK (
        pg_catalog.octet_length(source_identity) BETWEEN 1 AND 256
    ),
    source_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(source_sha256)=32),
    source_document jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(source_document)='object'),
    target_kind text NOT NULL CHECK (target_kind IN (
        'discard','installation','binding','dependency',
        'menu_definition','menu_version','menu_item'
    )),
    target_identity text NOT NULL CHECK (
        pg_catalog.octet_length(target_identity) BETWEEN 1 AND 256
    ),
    target_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(target_sha256)=32),
    target_record jsonb NOT NULL,
    removed boolean NOT NULL,
    source_xid xid8 NOT NULL,
    commit_lsn pg_lsn NOT NULL,
    batch_ordinal integer NOT NULL CHECK (batch_ordinal BETWEEN 1 AND 2147483),
    change_ordinal smallint NOT NULL CHECK (change_ordinal BETWEEN 1 AND 1000),
    recorded_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id,run_id,target_kind,target_identity),
    CHECK (source_sha256=pg_catalog.sha256(
        pg_catalog.convert_to(source_document::text,'UTF8')
    )),
    CHECK (target_sha256=pg_catalog.sha256(
        pg_catalog.convert_to(target_record::text,'UTF8')
    )),
    CHECK ((target_kind='discard')=(target_record='null'::jsonb))
);

CREATE TABLE catalog_v1_adoption_receipts (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    cutover_generation bigint NOT NULL CHECK (cutover_generation>0),
    plan_sha256 text NOT NULL CHECK (plan_sha256~'^[0-9a-f]{64}$'),
    transform_sha256 text NOT NULL CHECK (transform_sha256~'^[0-9a-f]{64}$'),
    lease_manifest_sha256 text NOT NULL CHECK (
        lease_manifest_sha256~'^[0-9a-f]{64}$'
    ),
    final_lsn pg_lsn NOT NULL,
    source_count bigint NOT NULL CHECK (source_count>=0),
    source_root_sha256 text NOT NULL CHECK (source_root_sha256~'^[0-9a-f]{64}$'),
    target_count bigint NOT NULL CHECK (target_count>=0),
    target_root_sha256 text NOT NULL CHECK (target_root_sha256~'^[0-9a-f]{64}$'),
    blocker_sha256 text NOT NULL CHECK (blocker_sha256~'^[0-9a-f]{64}$'),
    consumer_sha256 text NOT NULL CHECK (consumer_sha256~'^[0-9a-f]{64}$'),
    evidence_sha256 text NOT NULL CHECK (evidence_sha256~'^[0-9a-f]{64}$'),
    acceptance_sha256 text NOT NULL CHECK (acceptance_sha256~'^[0-9a-f]{64}$'),
    recorded_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id,run_id),
    CHECK (date_trunc('microseconds',recorded_at)=recorded_at)
);

-- +goose StatementBegin
CREATE FUNCTION catalog_v1_adoption_receipt_immutable_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Catalog V1 adoption receipt is immutable';
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER catalog_v1_adoption_receipts_immutable
BEFORE UPDATE OR DELETE ON catalog_v1_adoption_receipts
FOR EACH STATEMENT EXECUTE FUNCTION catalog_v1_adoption_receipt_immutable_v1();
ALTER TABLE catalog_v1_adoption_receipts
    ENABLE ALWAYS TRIGGER catalog_v1_adoption_receipts_immutable;

-- PostgreSQL creates these tables without PUBLIC relation privileges.
REVOKE ALL ON FUNCTION catalog_v1_adoption_receipt_immutable_v1() FROM PUBLIC;
