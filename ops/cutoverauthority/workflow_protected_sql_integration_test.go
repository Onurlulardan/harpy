//go:build integration

package main

const workflowCatalogAssertSQL = `CREATE FUNCTION cutover_authority.assert_adopted_catalog_v1(
p_receipt_sha256 text) RETURNS void LANGUAGE plpgsql SECURITY DEFINER
SET search_path=pg_catalog AS $body$
BEGIN
 IF NOT EXISTS (SELECT 1 FROM cutover_authority.authority_adoption_inventory
                WHERE receipt_sha256=p_receipt_sha256) THEN
  RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='receipt missing';
 END IF;
 IF pg_catalog.pg_get_userbyid((SELECT nspowner FROM pg_catalog.pg_namespace
    WHERE nspname='cutover_authority'))<>'harpy_cutover_authority_owner' OR
    EXISTS (SELECT 1 FROM pg_catalog.pg_class AS relation
      JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
      WHERE namespace.nspname='cutover_authority' AND relation.relkind IN ('r','p')
        AND pg_catalog.pg_get_userbyid(relation.relowner)<>'harpy_cutover_authority_owner') OR
    EXISTS (SELECT 1 FROM pg_catalog.pg_proc AS procedure
      JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure.pronamespace
      WHERE namespace.nspname='cutover_authority'
        AND pg_catalog.pg_get_userbyid(procedure.proowner)<>'harpy_cutover_authority_owner') THEN
  RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='catalog owner drift';
 END IF;
END;$body$`

const workflowStateHashSQL = `CREATE FUNCTION cutover_authority.finalization_state_sha256_v1(
p_operation_id uuid,p_run_id uuid) RETURNS text LANGUAGE sql SECURITY DEFINER
SET search_path=pg_catalog AS $body$ SELECT pg_catalog.encode(pg_catalog.sha256(
pg_catalog.convert_to(p_operation_id::text||':'||p_run_id::text||':'||
 receipt.seal_receipt_sha256||':'||receipt.cleanup_receipt_sha256,'UTF8')||
 receipt.seal_receipt_document||receipt.cleanup_receipt_document),'hex')
FROM platform.cutover_credential_receipts AS receipt
WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id $body$`

const workflowRecordCredentialReceiptsSQL = `CREATE FUNCTION
cutover_authority.record_credential_receipts_v1(
p_operation_id uuid,p_run_id uuid,p_seal_sha text,p_seal_doc bytea,
p_cleanup_sha text,p_cleanup_doc bytea) RETURNS void LANGUAGE plpgsql
SECURITY DEFINER SET search_path=pg_catalog AS $body$
BEGIN
 IF SESSION_USER<>'harpy_cluster_admin' THEN
  RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='recorder denied';
 END IF;
 IF p_seal_sha<>pg_catalog.encode(pg_catalog.sha256(p_seal_doc),'hex') OR
    p_cleanup_sha<>pg_catalog.encode(pg_catalog.sha256(p_cleanup_doc),'hex') THEN
  RAISE EXCEPTION USING ERRCODE='22000',MESSAGE='receipt digest differs';
 END IF;
 INSERT INTO platform.cutover_credential_receipts(
  operation_id,run_id,seal_receipt_sha256,seal_receipt_document,
  cleanup_receipt_sha256,cleanup_receipt_document)
 VALUES(p_operation_id,p_run_id,p_seal_sha,p_seal_doc,p_cleanup_sha,p_cleanup_doc)
 ON CONFLICT(operation_id,run_id) DO NOTHING;
 IF NOT EXISTS (SELECT 1 FROM platform.cutover_credential_receipts AS receipt
  WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id
    AND receipt.seal_receipt_sha256=p_seal_sha
    AND receipt.seal_receipt_document=p_seal_doc
    AND receipt.cleanup_receipt_sha256=p_cleanup_sha
    AND receipt.cleanup_receipt_document=p_cleanup_doc) THEN
  RAISE EXCEPTION USING ERRCODE='22000',MESSAGE='credential receipt replay differs';
 END IF;
END;$body$`

const workflowObserveAuthoritySQL = `CREATE FUNCTION cutover_authority.observe_finalizer_authority_v1(
p_operation_id uuid,p_run_id uuid) RETURNS TABLE(
finalization_state_sha256 text,finalization_procedure_oid oid,
finalization_procedure_signature text,database_name text,database_origin text,
system_identifier text,database_oid oid,operation_state text,
completed_eligible boolean,abort_eligible boolean,
completed_tenant_catalog_sha256 text,completed_tenant_requirement_count integer,
finalizer_valid_until timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog AS $body$
DECLARE v_role pg_catalog.pg_roles%ROWTYPE;
BEGIN
 IF SESSION_USER<>'harpy_cluster_admin' THEN
  RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='observer denied';
 END IF;
 SELECT role.* INTO STRICT v_role FROM pg_catalog.pg_roles AS role
 WHERE role.rolname='harpy_cutover_finalizer' AND role.rolcanlogin
   AND NOT role.rolinherit AND NOT role.rolsuper AND NOT role.rolcreatedb
   AND NOT role.rolcreaterole AND NOT role.rolreplication AND NOT role.rolbypassrls
   AND role.rolconnlimit=1 AND role.rolpassword='********'
   AND role.rolvaliduntil>pg_catalog.clock_timestamp()+interval '15 seconds'
   AND role.rolvaliduntil<=pg_catalog.clock_timestamp()+interval '120 seconds'
   AND NOT EXISTS(SELECT 1 FROM pg_catalog.pg_auth_members AS membership
     WHERE membership.roleid=role.oid OR membership.member=role.oid);
 RETURN QUERY SELECT cutover_authority.finalization_state_sha256_v1(
   p_operation_id,p_run_id),
   'cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)'::regprocedure::oid,
   'cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)',
   current_database()::text,'fresh'::text,control.system_identifier::text,
   database.oid,
   CASE WHEN p_operation_id='44444444-4444-4444-8444-444444444446'::uuid
        THEN 'quiescing' ELSE 'blocked' END,
   p_operation_id='44444444-4444-4444-8444-444444444446'::uuid,
   p_operation_id<>'44444444-4444-4444-8444-444444444446'::uuid,
   pg_catalog.repeat('a',64),1,
   v_role.rolvaliduntil
 FROM pg_catalog.pg_control_system() AS control
 JOIN pg_catalog.pg_database AS database ON database.datname=current_database();
END;$body$`

const workflowValidateEnvelopeSQL = `CREATE FUNCTION cutover_authority.validate_finalization_envelope_v1(
uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $body$ BEGIN RETURN; END;$body$`

const workflowAssertCompletedSQL = `CREATE FUNCTION cutover_authority.assert_completed_evidence_v1(
uuid,uuid,bytea) RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $body$ BEGIN RETURN; END;$body$`

const workflowAssertTenantSQL = `CREATE FUNCTION cutover_authority.assert_tenant_completion_v1(
uuid,uuid) RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $body$ BEGIN RETURN; END;$body$`

const workflowAssertTerminalSQL = `CREATE FUNCTION cutover_authority.assert_terminal_evidence_v1(
uuid,uuid,text,bytea) RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $body$ BEGIN RETURN; END;$body$`

const workflowAuthorizeSQL = `CREATE FUNCTION cutover_authority.authorize_finalization_v1(
p_operation_id uuid,p_run_id uuid,p_lease_id uuid,
p_finalizer_sha text,p_finalizer_doc bytea,p_adoption_sha text,p_adoption_doc bytea,
p_seal_sha text,p_seal_doc bytea,p_cleanup_sha text,p_cleanup_doc bytea,
p_authority_sha text,p_authority_doc bytea,p_terminal_state text,
p_terminal_sha text,p_terminal_doc bytea) RETURNS void LANGUAGE plpgsql SECURITY DEFINER
SET search_path=pg_catalog AS $body$
DECLARE v_payload text;
BEGIN
 IF SESSION_USER<>'harpy_cluster_admin' THEN
  RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='authorizer denied';
 END IF;
 IF p_finalizer_sha<>pg_catalog.encode(pg_catalog.sha256(p_finalizer_doc),'hex') OR
    p_adoption_sha<>pg_catalog.encode(pg_catalog.sha256(p_adoption_doc),'hex') OR
    p_seal_sha<>pg_catalog.encode(pg_catalog.sha256(p_seal_doc),'hex') OR
    p_cleanup_sha<>pg_catalog.encode(pg_catalog.sha256(p_cleanup_doc),'hex') OR
    p_authority_sha<>pg_catalog.encode(pg_catalog.sha256(p_authority_doc),'hex') OR
    p_terminal_sha<>pg_catalog.encode(pg_catalog.sha256(p_terminal_doc),'hex') THEN
  RAISE EXCEPTION USING ERRCODE='22000',MESSAGE='document digest differs';
 END IF;
 v_payload:=pg_catalog.encode(pg_catalog.sha256(p_finalizer_doc||p_adoption_doc||
  p_seal_doc||p_cleanup_doc||p_authority_doc||p_terminal_doc||pg_catalog.convert_to(
  p_terminal_state||p_finalizer_sha||p_adoption_sha||p_seal_sha||p_cleanup_sha||
  p_authority_sha||p_terminal_sha,'UTF8')),'hex');
 INSERT INTO cutover_authority.finalization_authorizations(
  operation_id,run_id,finalizer_lease_id,payload_sha256)
 VALUES(p_operation_id,p_run_id,p_lease_id,v_payload) ON CONFLICT DO NOTHING;
 IF NOT EXISTS (SELECT 1 FROM cutover_authority.finalization_authorizations
  WHERE operation_id=p_operation_id AND run_id=p_run_id
    AND finalizer_lease_id=p_lease_id AND payload_sha256=v_payload) THEN
  RAISE EXCEPTION USING ERRCODE='22000',MESSAGE='authorization replay differs';
 END IF;
END;$body$`

const workflowConsumeSQL = `CREATE FUNCTION cutover_authority.consume_finalization_authorization_v1(
uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $body$ BEGIN RETURN; END;$body$`

const workflowFinalizeSQL = `CREATE FUNCTION cutover_authority.finalize_operation_v1(
p_operation_id uuid,p_run_id uuid,p_lease_id uuid,
text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,
p_terminal_state text,p_terminal_sha text,bytea)
RETURNS TABLE(terminal_state text,terminal_evidence_sha256 text)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $body$ SELECT p_terminal_state,p_terminal_sha $body$`

const workflowVerifySQL = `CREATE FUNCTION cutover_authority.verify_finalization_receipt_v1(
uuid,uuid,uuid,text,text) RETURNS boolean LANGUAGE sql SECURITY DEFINER
SET search_path=pg_catalog AS $body$ SELECT true $body$`
