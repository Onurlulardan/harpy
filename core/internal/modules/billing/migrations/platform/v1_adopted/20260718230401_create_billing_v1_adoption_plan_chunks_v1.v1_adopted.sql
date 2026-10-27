-- +goose Up
CREATE TABLE platform.billing_v1_adoption_plan_chunks (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    chunk_ordinal integer NOT NULL,
    chunk_count integer NOT NULL,
    chunk_sha256 bytea NOT NULL,
    chunk_bytes bytea NOT NULL,
    source_xid xid8 NOT NULL,
    commit_lsn pg_lsn NOT NULL,
    batch_ordinal integer NOT NULL,
    change_ordinal integer NOT NULL,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id, run_id, chunk_ordinal),
    CONSTRAINT billing_v1_adoption_plan_chunks_scope_check CHECK (
        scope_kind IN ('global', 'tenant') AND
        scope_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT billing_v1_adoption_plan_chunks_position_check CHECK (
        chunk_count BETWEEN 1 AND 65536 AND
        chunk_ordinal BETWEEN 1 AND chunk_count AND
        batch_ordinal BETWEEN 1 AND 2147483 AND
        change_ordinal BETWEEN 1 AND 1000
    ),
    CONSTRAINT billing_v1_adoption_plan_chunks_bytes_check CHECK (
        pg_catalog.octet_length(chunk_bytes) BETWEEN 1 AND 6291456 AND
        pg_catalog.octet_length(chunk_sha256) = 32 AND
        chunk_sha256 = pg_catalog.sha256(chunk_bytes)
    )
);

-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_v1_adoption_plan_chunk_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP <> 'INSERT' OR
       SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       NOT EXISTS (
           SELECT 1 FROM platform.cutover_operations AS operation
           WHERE operation.operation_id = NEW.operation_id
             AND operation.run_id = NEW.run_id
             AND operation.scope_kind = NEW.scope_kind
             AND operation.scope_id = NEW.scope_id
             AND operation.state IN ('copying', 'shadow', 'quiescing')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 adoption plan chunk mutation denied';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_guard_v1_adoption_plan_chunk_v1()
    FROM PUBLIC;
CREATE TRIGGER billing_v1_adoption_plan_chunks_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_v1_adoption_plan_chunks
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_v1_adoption_plan_chunk_v1();
ALTER TABLE platform.billing_v1_adoption_plan_chunks
    ENABLE ALWAYS TRIGGER billing_v1_adoption_plan_chunks_guard;
