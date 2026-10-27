-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.authorize_finalization_v1(
    p_operation_id uuid, p_run_id uuid, p_finalizer_lease_id uuid,
    p_finalizer_lease_sha256 text, p_finalizer_lease_document bytea, p_adoption_lease_sha256 text,
    p_adoption_lease_document bytea, p_seal_receipt_sha256 text, p_seal_receipt_document bytea,
    p_cleanup_receipt_sha256 text, p_cleanup_receipt_document bytea, p_authority_receipt_sha256 text,
    p_authority_receipt_document bytea, p_terminal_state text, p_terminal_evidence_sha256 text,
    p_terminal_evidence_document bytea
)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_finalizer jsonb;
    v_state_sha256 text;
    v_finalizer_expires_at timestamptz;
    v_authorizer_valid_until timestamptz;
BEGIN
    IF SESSION_USER <> 'forgecrud_cluster_admin' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='cutover finalization authorizer denied';
    END IF;
    PERFORM cutover_authority.validate_finalization_envelope_v1(
        p_operation_id,p_run_id,p_finalizer_lease_id,p_finalizer_lease_sha256,p_finalizer_lease_document,
        p_adoption_lease_sha256,p_adoption_lease_document,p_seal_receipt_sha256,p_seal_receipt_document,
        p_cleanup_receipt_sha256,p_cleanup_receipt_document,p_authority_receipt_sha256,p_authority_receipt_document,
        p_terminal_state,p_terminal_evidence_sha256,p_terminal_evidence_document);
    PERFORM 1 FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id FOR UPDATE;
    IF EXISTS (SELECT 1 FROM cutover_authority.finalization_receipts AS receipt
        WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover operation is already finalized';
    END IF;
    IF EXISTS (SELECT 1 FROM cutover_authority.finalization_authorizations AS authorization_row
        WHERE authorization_row.operation_id=p_operation_id AND authorization_row.run_id=p_run_id
          AND authorization_row.finalizer_lease_id<>p_finalizer_lease_id
          AND authorization_row.finalizer_expires_at>pg_catalog.clock_timestamp()) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover finalization authorization attempt is active';
    END IF;
    v_finalizer:=pg_catalog.convert_from(p_finalizer_lease_document,'UTF8')::jsonb;
    v_finalizer_expires_at:=(v_finalizer->>'expires_at')::timestamptz;
    SELECT role.rolvaliduntil INTO STRICT v_authorizer_valid_until FROM pg_catalog.pg_roles AS role
    WHERE role.rolname='forgecrud_cluster_admin' AND role.rolcanlogin AND NOT role.rolinherit
      AND NOT role.rolsuper AND role.rolcreatedb AND role.rolcreaterole AND NOT role.rolreplication
      AND NOT role.rolbypassrls AND role.rolconnlimit=1 AND role.rolvaliduntil>pg_catalog.clock_timestamp()
      AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
          WHERE membership.roleid=role.oid OR membership.member=role.oid);
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
        WHERE role.rolname='forgecrud_cutover_finalizer' AND role.rolcanlogin AND NOT role.rolinherit
          AND NOT role.rolsuper AND NOT role.rolcreatedb AND NOT role.rolcreaterole
          AND NOT role.rolreplication AND NOT role.rolbypassrls AND role.rolconnlimit=1
          AND role.rolvaliduntil IS NOT DISTINCT FROM v_finalizer_expires_at
          AND role.rolvaliduntil>pg_catalog.clock_timestamp()
          AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
              WHERE membership.roleid=role.oid OR membership.member=role.oid)) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover finalizer role lease is not active';
    END IF;
    v_state_sha256:=cutover_authority.finalization_state_sha256_v1(p_operation_id,p_run_id);
    IF v_state_sha256 IS DISTINCT FROM v_finalizer->>'finalization_state_sha256' OR
       v_finalizer_expires_at<=pg_catalog.clock_timestamp() OR
       v_finalizer_expires_at>v_authorizer_valid_until THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover finalization authorization state changed';
    END IF;
    INSERT INTO cutover_authority.finalization_authorizations (
        operation_id,run_id,finalizer_lease_id,finalizer_lease_sha256,finalizer_lease_document,
        adoption_lease_sha256,adoption_lease_document,seal_receipt_sha256,seal_receipt_document,
        cleanup_receipt_sha256,cleanup_receipt_document,authority_receipt_sha256,authority_receipt_document,
        terminal_state,terminal_evidence_sha256,terminal_evidence_document,finalization_state_sha256,
        finalizer_expires_at,authorizer_valid_until)
    VALUES (p_operation_id,p_run_id,p_finalizer_lease_id,p_finalizer_lease_sha256,p_finalizer_lease_document,
        p_adoption_lease_sha256,p_adoption_lease_document,p_seal_receipt_sha256,p_seal_receipt_document,
        p_cleanup_receipt_sha256,p_cleanup_receipt_document,p_authority_receipt_sha256,p_authority_receipt_document,
        p_terminal_state,p_terminal_evidence_sha256,p_terminal_evidence_document,v_state_sha256,
        v_finalizer_expires_at,v_authorizer_valid_until)
    ON CONFLICT DO NOTHING;
    IF NOT FOUND AND NOT EXISTS (SELECT 1 FROM cutover_authority.finalization_authorizations AS authorization_row
        WHERE authorization_row.operation_id=p_operation_id AND authorization_row.run_id=p_run_id
          AND authorization_row.finalizer_lease_id=p_finalizer_lease_id
          AND authorization_row.finalizer_lease_sha256=p_finalizer_lease_sha256
          AND authorization_row.finalizer_lease_document=p_finalizer_lease_document
          AND authorization_row.adoption_lease_sha256=p_adoption_lease_sha256
          AND authorization_row.adoption_lease_document=p_adoption_lease_document
          AND authorization_row.seal_receipt_sha256=p_seal_receipt_sha256
          AND authorization_row.seal_receipt_document=p_seal_receipt_document
          AND authorization_row.cleanup_receipt_sha256=p_cleanup_receipt_sha256
          AND authorization_row.cleanup_receipt_document=p_cleanup_receipt_document
          AND authorization_row.authority_receipt_sha256=p_authority_receipt_sha256
          AND authorization_row.authority_receipt_document=p_authority_receipt_document
          AND authorization_row.terminal_state=p_terminal_state
          AND authorization_row.terminal_evidence_sha256=p_terminal_evidence_sha256
          AND authorization_row.terminal_evidence_document=p_terminal_evidence_document
          AND authorization_row.finalization_state_sha256=v_state_sha256
          AND authorization_row.finalizer_expires_at=v_finalizer_expires_at
          AND authorization_row.authorizer_valid_until=v_authorizer_valid_until) THEN
        RAISE EXCEPTION USING ERRCODE='22000', MESSAGE='cutover finalization authorization replay differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.authorize_finalization_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION cutover_authority.authorize_finalization_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea) TO forgecrud_cluster_admin;
