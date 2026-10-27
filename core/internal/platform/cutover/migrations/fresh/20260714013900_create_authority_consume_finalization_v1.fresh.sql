-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.consume_finalization_authorization_v1(
    p_operation_id uuid, p_run_id uuid, p_finalizer_lease_id uuid,
    p_finalizer_lease_sha256 text, p_finalizer_lease_document bytea, p_adoption_lease_sha256 text,
    p_adoption_lease_document bytea, p_seal_receipt_sha256 text, p_seal_receipt_document bytea,
    p_cleanup_receipt_sha256 text, p_cleanup_receipt_document bytea, p_authority_receipt_sha256 text,
    p_authority_receipt_document bytea, p_terminal_state text, p_terminal_evidence_sha256 text,
    p_terminal_evidence_document bytea
)
RETURNS TABLE (already_finalized boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_authorization cutover_authority.finalization_authorizations%ROWTYPE;
    v_receipt cutover_authority.finalization_receipts%ROWTYPE;
    v_finalizer jsonb;
BEGIN
    IF SESSION_USER <> 'forgecrud_cutover_finalizer' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='cutover finalization consumer denied';
    END IF;
    SELECT authorization_row.* INTO STRICT v_authorization
    FROM cutover_authority.finalization_authorizations AS authorization_row
    WHERE authorization_row.operation_id=p_operation_id AND authorization_row.run_id=p_run_id
      AND authorization_row.finalizer_lease_id=p_finalizer_lease_id FOR UPDATE;
    IF v_authorization.finalizer_lease_id IS DISTINCT FROM p_finalizer_lease_id OR
       v_authorization.finalizer_lease_sha256 IS DISTINCT FROM p_finalizer_lease_sha256 OR
       v_authorization.finalizer_lease_document IS DISTINCT FROM p_finalizer_lease_document OR
       v_authorization.adoption_lease_sha256 IS DISTINCT FROM p_adoption_lease_sha256 OR
       v_authorization.adoption_lease_document IS DISTINCT FROM p_adoption_lease_document OR
       v_authorization.seal_receipt_sha256 IS DISTINCT FROM p_seal_receipt_sha256 OR
       v_authorization.seal_receipt_document IS DISTINCT FROM p_seal_receipt_document OR
       v_authorization.cleanup_receipt_sha256 IS DISTINCT FROM p_cleanup_receipt_sha256 OR
       v_authorization.cleanup_receipt_document IS DISTINCT FROM p_cleanup_receipt_document OR
       v_authorization.authority_receipt_sha256 IS DISTINCT FROM p_authority_receipt_sha256 OR
       v_authorization.authority_receipt_document IS DISTINCT FROM p_authority_receipt_document OR
       v_authorization.terminal_state IS DISTINCT FROM p_terminal_state OR
       v_authorization.terminal_evidence_sha256 IS DISTINCT FROM p_terminal_evidence_sha256 OR
       v_authorization.terminal_evidence_document IS DISTINCT FROM p_terminal_evidence_document OR
       pg_catalog.encode(pg_catalog.sha256(p_finalizer_lease_document),'hex') IS DISTINCT FROM p_finalizer_lease_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(p_adoption_lease_document),'hex') IS DISTINCT FROM p_adoption_lease_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(p_seal_receipt_document),'hex') IS DISTINCT FROM p_seal_receipt_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(p_cleanup_receipt_document),'hex') IS DISTINCT FROM p_cleanup_receipt_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(p_authority_receipt_document),'hex') IS DISTINCT FROM p_authority_receipt_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(p_terminal_evidence_document),'hex') IS DISTINCT FROM p_terminal_evidence_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='22000', MESSAGE='cutover finalization authorization differs';
    END IF;
    SELECT receipt.* INTO v_receipt FROM cutover_authority.finalization_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id;
    IF FOUND THEN
        IF v_receipt.finalizer_lease_id IS DISTINCT FROM p_finalizer_lease_id OR
           v_receipt.finalizer_lease_sha256 IS DISTINCT FROM p_finalizer_lease_sha256 OR
           v_receipt.finalizer_lease_document IS DISTINCT FROM p_finalizer_lease_document OR
           v_receipt.adoption_lease_sha256 IS DISTINCT FROM p_adoption_lease_sha256 OR
           v_receipt.adoption_lease_document IS DISTINCT FROM p_adoption_lease_document OR
           v_receipt.seal_receipt_sha256 IS DISTINCT FROM p_seal_receipt_sha256 OR
           v_receipt.seal_receipt_document IS DISTINCT FROM p_seal_receipt_document OR
           v_receipt.cleanup_receipt_sha256 IS DISTINCT FROM p_cleanup_receipt_sha256 OR
           v_receipt.cleanup_receipt_document IS DISTINCT FROM p_cleanup_receipt_document OR
           v_receipt.authority_receipt_sha256 IS DISTINCT FROM p_authority_receipt_sha256 OR
           v_receipt.authority_receipt_document IS DISTINCT FROM p_authority_receipt_document OR
           v_receipt.terminal_state IS DISTINCT FROM p_terminal_state OR
           v_receipt.terminal_evidence_sha256 IS DISTINCT FROM p_terminal_evidence_sha256 OR
           v_receipt.terminal_evidence_document IS DISTINCT FROM p_terminal_evidence_document OR
           v_receipt.finalization_state_sha256 IS DISTINCT FROM v_authorization.finalization_state_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='22000', MESSAGE='cutover finalization receipt replay differs';
        END IF;
        RETURN QUERY SELECT true;
        RETURN;
    END IF;
    v_finalizer:=pg_catalog.convert_from(p_finalizer_lease_document,'UTF8')::jsonb;
    IF v_authorization.finalizer_expires_at<=pg_catalog.clock_timestamp() OR
       v_authorization.authorizer_valid_until<=pg_catalog.clock_timestamp() OR
       v_finalizer->>'finalization_state_sha256' IS DISTINCT FROM v_authorization.finalization_state_sha256 OR
       (v_finalizer->>'expires_at')::timestamptz IS DISTINCT FROM v_authorization.finalizer_expires_at OR
       NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
          WHERE role.rolname='forgecrud_cutover_finalizer' AND role.rolcanlogin
            AND NOT role.rolinherit AND NOT role.rolsuper AND NOT role.rolcreatedb
            AND NOT role.rolcreaterole AND NOT role.rolreplication AND NOT role.rolbypassrls
            AND role.rolconnlimit=1 AND role.rolvaliduntil IS NOT DISTINCT FROM v_authorization.finalizer_expires_at
            AND role.rolvaliduntil>pg_catalog.clock_timestamp()
            AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
                WHERE membership.roleid=role.oid OR membership.member=role.oid)) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover finalizer role lease is not active';
    END IF;
    RETURN QUERY SELECT false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.consume_finalization_authorization_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea) FROM PUBLIC;
