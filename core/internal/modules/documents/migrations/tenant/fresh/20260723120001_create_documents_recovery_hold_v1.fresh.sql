-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_recovery_begin_hold_v1(
    p_recovery_set_id uuid,
    p_environment_id uuid,
    p_tenant_id uuid,
    p_tenant_set_sha256 bytea
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_existing documents_dr_holds%ROWTYPE;
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_backup'
       OR p_recovery_set_id IS NULL OR p_environment_id IS NULL
       OR p_tenant_id IS NULL
       OR pg_catalog.octet_length(p_tenant_set_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'documents recovery hold denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'documents:dr-hold:' || pg_catalog.current_schema(), 0
        )
    );
    IF NOT EXISTS (
        SELECT 1 FROM tenancy_schema_identity AS identity
        WHERE identity.singleton AND identity.tenant_id = p_tenant_id
          AND identity.schema_name = pg_catalog.current_schema()
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'documents recovery tenant mismatch';
    END IF;
    SELECT * INTO v_existing
    FROM documents_dr_holds
    WHERE recovery_set_id = p_recovery_set_id;
    IF FOUND THEN
        IF v_existing.environment_id <> p_environment_id
           OR v_existing.tenant_id <> p_tenant_id
           OR v_existing.tenant_set_sha256 <> p_tenant_set_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'documents recovery hold conflicts';
        END IF;
        RETURN v_existing.lifecycle_state;
    END IF;
    IF EXISTS (
        SELECT 1 FROM documents_dr_holds AS hold
        WHERE hold.lifecycle_state IN ('HOLDING','COPYING','VERIFIED')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'documents recovery hold is already active';
    END IF;
    IF EXISTS (
        SELECT 1 FROM documents_blob_deletions AS deletion
        WHERE deletion.target_kind IN ('VERSION','ARCHIVE')
          AND deletion.lifecycle_state IN ('CLAIMED','RECONCILING')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'documents durable deletion is in flight';
    END IF;
    INSERT INTO documents_dr_holds(
        recovery_set_id, environment_id, tenant_id,
        tenant_set_sha256, lifecycle_state,
        started_at, updated_at
    ) VALUES (
        p_recovery_set_id, p_environment_id, p_tenant_id,
        p_tenant_set_sha256, 'HOLDING',
        v_now, v_now
    );
    RETURN 'HOLDING';
END;
$function$;
-- +goose StatementEnd
