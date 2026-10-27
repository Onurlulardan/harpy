-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.finalize_operation_v1(
    p_operation_id uuid, p_run_id uuid, p_finalizer_lease_id uuid,
    p_finalizer_lease_sha256 text, p_finalizer_lease_document bytea, p_adoption_lease_sha256 text,
    p_adoption_lease_document bytea, p_seal_receipt_sha256 text, p_seal_receipt_document bytea,
    p_cleanup_receipt_sha256 text, p_cleanup_receipt_document bytea, p_authority_receipt_sha256 text,
    p_authority_receipt_document bytea, p_terminal_state text, p_terminal_evidence_sha256 text,
    p_terminal_evidence_document bytea
)
RETURNS TABLE (terminal_state text, terminal_evidence_sha256 text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_already_finalized boolean;
    v_authorization cutover_authority.finalization_authorizations%ROWTYPE;
    v_operation platform.cutover_operations%ROWTYPE;
    v_finalizer jsonb;
    v_state_sha256 text;
    v_catalog_sha256 text;
    v_catalog_document bytea;
    v_requirement_count integer;
    v_target_count integer;
    v_updated_state text;
BEGIN
    IF SESSION_USER <> 'harpy_cutover_finalizer' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='cutover finalizer denied';
    END IF;
    SELECT result.already_finalized INTO STRICT v_already_finalized
    FROM cutover_authority.consume_finalization_authorization_v1(
        p_operation_id,p_run_id,p_finalizer_lease_id,p_finalizer_lease_sha256,p_finalizer_lease_document,
        p_adoption_lease_sha256,p_adoption_lease_document,p_seal_receipt_sha256,p_seal_receipt_document,
        p_cleanup_receipt_sha256,p_cleanup_receipt_document,p_authority_receipt_sha256,p_authority_receipt_document,
        p_terminal_state,p_terminal_evidence_sha256,p_terminal_evidence_document) AS result;
    IF v_already_finalized THEN
        RETURN QUERY SELECT receipt.terminal_state,receipt.terminal_evidence_sha256
        FROM cutover_authority.finalization_receipts AS receipt
        WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id
          AND receipt.finalizer_lease_id=p_finalizer_lease_id;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover finalized operation differs';
        END IF;
        RETURN;
    END IF;
    PERFORM cutover_authority.validate_finalization_envelope_v1(
        p_operation_id,p_run_id,p_finalizer_lease_id,p_finalizer_lease_sha256,p_finalizer_lease_document,
        p_adoption_lease_sha256,p_adoption_lease_document,p_seal_receipt_sha256,p_seal_receipt_document,
        p_cleanup_receipt_sha256,p_cleanup_receipt_document,p_authority_receipt_sha256,p_authority_receipt_document,
        p_terminal_state,p_terminal_evidence_sha256,p_terminal_evidence_document);
    SELECT authorization_row.* INTO STRICT v_authorization
    FROM cutover_authority.finalization_authorizations AS authorization_row
    WHERE authorization_row.operation_id=p_operation_id AND authorization_row.run_id=p_run_id
      AND authorization_row.finalizer_lease_id=p_finalizer_lease_id FOR UPDATE;
    SELECT operation.* INTO STRICT v_operation FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id;
    IF p_terminal_state='completed' THEN
        SELECT catalog.catalog_sha256,catalog.canonical_document,catalog.requirement_count,catalog.target_count
        INTO STRICT v_catalog_sha256,v_catalog_document,v_requirement_count,v_target_count
        FROM platform.release_generations AS release_row
        JOIN platform.tenant_migration_requirement_catalogs AS catalog
          ON catalog.migration_run_id=release_row.migration_run_id
         AND catalog.release_id=release_row.release_id
         AND catalog.environment_id=release_row.environment_id
         AND catalog.database_origin=release_row.database_origin
        WHERE release_row.release_id=v_operation.release_id
          AND release_row.environment_id=v_operation.environment_id;
    END IF;
    v_finalizer:=pg_catalog.convert_from(p_finalizer_lease_document,'UTF8')::jsonb;
    v_state_sha256:=cutover_authority.finalization_state_sha256_v1(p_operation_id,p_run_id);
    IF v_state_sha256 IS DISTINCT FROM v_authorization.finalization_state_sha256 OR
       v_state_sha256 IS DISTINCT FROM v_finalizer->>'finalization_state_sha256' OR
       v_authorization.finalizer_expires_at<=pg_catalog.clock_timestamp() OR
       v_authorization.authorizer_valid_until<=pg_catalog.clock_timestamp() OR
       NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
          WHERE role.rolname='harpy_cutover_finalizer' AND role.rolcanlogin AND NOT role.rolinherit
            AND NOT role.rolsuper AND NOT role.rolcreatedb AND NOT role.rolcreaterole
            AND NOT role.rolreplication AND NOT role.rolbypassrls AND role.rolconnlimit=1
            AND role.rolvaliduntil IS NOT DISTINCT FROM v_authorization.finalizer_expires_at
            AND role.rolvaliduntil>pg_catalog.clock_timestamp() AND NOT EXISTS
              (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
               WHERE membership.roleid=role.oid OR membership.member=role.oid)) OR
       NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
          WHERE role.rolname='harpy_cluster_admin' AND role.rolcanlogin AND NOT role.rolinherit
            AND NOT role.rolsuper AND role.rolcreatedb AND role.rolcreaterole
            AND NOT role.rolreplication AND NOT role.rolbypassrls AND role.rolconnlimit=1
            AND role.rolvaliduntil IS NOT DISTINCT FROM v_authorization.authorizer_valid_until
            AND role.rolvaliduntil>pg_catalog.clock_timestamp() AND NOT EXISTS
              (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
               WHERE membership.roleid=role.oid OR membership.member=role.oid)) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover finalization state or role lease changed';
    END IF;
    INSERT INTO cutover_authority.finalization_receipts (
        operation_id,run_id,finalizer_lease_id,finalizer_lease_sha256,finalizer_lease_document,
        adoption_lease_sha256,adoption_lease_document,seal_receipt_sha256,seal_receipt_document,
        cleanup_receipt_sha256,cleanup_receipt_document,authority_receipt_sha256,authority_receipt_document,
        terminal_state,terminal_evidence_sha256,terminal_evidence_document,finalization_state_sha256,
        tenant_requirement_catalog_sha256,tenant_requirement_catalog_document,
        tenant_requirement_count,tenant_target_count)
    VALUES (p_operation_id,p_run_id,p_finalizer_lease_id,p_finalizer_lease_sha256,p_finalizer_lease_document,
        p_adoption_lease_sha256,p_adoption_lease_document,p_seal_receipt_sha256,p_seal_receipt_document,
        p_cleanup_receipt_sha256,p_cleanup_receipt_document,p_authority_receipt_sha256,p_authority_receipt_document,
        p_terminal_state,p_terminal_evidence_sha256,p_terminal_evidence_document,v_state_sha256,
        v_catalog_sha256,v_catalog_document,v_requirement_count,v_target_count);
    UPDATE platform.cutover_operations AS operation
    SET state=p_terminal_state,blocked_reason=NULL,updated_at=pg_catalog.clock_timestamp()
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id AND
      ((p_terminal_state='completed' AND operation.state='quiescing') OR
       (p_terminal_state IN ('failed','cancelled') AND operation.state IN ('planned','copying','shadow','quiescing','blocked')))
    RETURNING operation.state INTO STRICT v_updated_state;
    RETURN QUERY SELECT v_updated_state,p_terminal_evidence_sha256;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea) TO harpy_cutover_finalizer;
