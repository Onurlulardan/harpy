-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_prepare_batch_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_scope_kind text,
    p_scope_id uuid,
    p_relation_ordinal smallint,
    p_checkpoint_key text,
    p_batch_ordinal integer,
    p_source_commit_lsn pg_lsn,
    p_source_xid xid8,
    p_batch_sha256 text,
    p_lease_manifest_sha256 text
)
RETURNS TABLE (
    procedure_contract_id text,
    procedure_oid bigint,
    procedure_signature text,
    already_applied boolean,
    applied_count bigint,
    key_state_sha256 text,
    state_sha256 text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE;
    v_procedure platform.cutover_operation_procedures%ROWTYPE;
    v_receipt platform.cutover_batch_receipts%ROWTYPE;
    v_live_signature text;
BEGIN
    IF SESSION_USER <> 'harpy_adopter' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'cutover batch prepare principal denied';
    END IF;
    SELECT operation.* INTO STRICT v_operation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id = p_operation_id AND operation.run_id = p_run_id
    FOR UPDATE;
    IF v_operation.lease_manifest_sha256 IS DISTINCT FROM p_lease_manifest_sha256 OR
       v_operation.scope_kind IS DISTINCT FROM p_scope_kind OR
       v_operation.scope_id IS DISTINCT FROM p_scope_id OR
       v_operation.state NOT IN ('copying', 'shadow', 'quiescing') OR
       v_operation.lease_expires_at <= pg_catalog.clock_timestamp() OR
       NOT EXISTS (SELECT 1 FROM platform.cutover_operation_checkpoint_keys AS checkpoint_key
           WHERE checkpoint_key.operation_id = p_operation_id
             AND checkpoint_key.run_id = p_run_id
             AND checkpoint_key.relation_ordinal = p_relation_ordinal
             AND checkpoint_key.checkpoint_key = p_checkpoint_key) OR
       NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
           WHERE role.rolname = 'harpy_adopter' AND NOT role.rolcanlogin
             AND NOT role.rolinherit AND NOT role.rolsuper AND NOT role.rolcreatedb
             AND NOT role.rolcreaterole AND NOT role.rolreplication AND NOT role.rolbypassrls
             AND role.rolconnlimit = 0
             AND role.rolvaliduntil IS NOT DISTINCT FROM v_operation.lease_expires_at
             AND role.rolvaliduntil > pg_catalog.clock_timestamp()
             AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
                 WHERE membership.roleid = role.oid OR membership.member = role.oid)) OR
       NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
           WHERE role.rolname = 'harpy_cutover_decoder' AND NOT role.rolcanlogin
             AND NOT role.rolinherit AND NOT role.rolsuper AND NOT role.rolcreatedb
             AND NOT role.rolcreaterole AND NOT role.rolreplication AND NOT role.rolbypassrls
             AND role.rolconnlimit = 0
             AND role.rolvaliduntil IS NOT DISTINCT FROM v_operation.lease_expires_at
             AND role.rolvaliduntil > pg_catalog.clock_timestamp()
             AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
                 WHERE membership.roleid = role.oid OR membership.member = role.oid)) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'cutover operation is not apply-ready';
    END IF;

    SELECT procedure_row.* INTO STRICT v_procedure
    FROM platform.cutover_operation_procedures AS procedure_row
    WHERE procedure_row.operation_id = p_operation_id
      AND procedure_row.run_id = p_run_id
      AND procedure_row.relation_ordinal = p_relation_ordinal
      AND procedure_row.procedure_kind = 'apply';
    SELECT pg_catalog.format(
        '%I.%I(%s)', namespace.nspname, procedure_row.proname,
        pg_catalog.replace(pg_catalog.oidvectortypes(procedure_row.proargtypes), ' ', '')
    ) INTO STRICT v_live_signature
    FROM pg_catalog.pg_proc AS procedure_row
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = procedure_row.pronamespace
    WHERE procedure_row.oid = v_procedure.procedure_oid::oid;
    IF v_live_signature <> v_procedure.procedure_signature THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'cutover apply procedure catalog differs';
    END IF;

    SELECT receipt.* INTO v_receipt
    FROM platform.cutover_batch_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id AND receipt.run_id = p_run_id
      AND receipt.relation_ordinal = p_relation_ordinal
      AND receipt.checkpoint_key = p_checkpoint_key
      AND receipt.batch_ordinal = p_batch_ordinal;
    IF FOUND AND (
        v_receipt.batch_sha256 <> p_batch_sha256 OR
        v_receipt.source_commit_lsn <> p_source_commit_lsn OR
        v_receipt.source_xid <> p_source_xid
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22000', MESSAGE = 'cutover batch idempotency tamper detected';
    END IF;

    RETURN QUERY SELECT
        v_procedure.procedure_contract_id,
        v_procedure.procedure_oid,
        v_procedure.procedure_signature,
        FOUND,
        CASE WHEN FOUND THEN v_receipt.applied_count ELSE NULL END,
        CASE WHEN FOUND THEN v_receipt.key_state_sha256 ELSE NULL END,
        CASE WHEN FOUND THEN v_receipt.state_sha256 ELSE NULL END;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_prepare_batch_v1(
    uuid, uuid, text, uuid, smallint, text, integer, pg_lsn, xid8, text, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.cutover_prepare_batch_v1(
    uuid, uuid, text, uuid, smallint, text, integer, pg_lsn, xid8, text, text
) TO harpy_adopter;
