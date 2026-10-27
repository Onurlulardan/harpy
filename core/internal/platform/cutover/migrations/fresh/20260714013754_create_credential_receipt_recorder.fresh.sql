-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.record_credential_receipts_v1(
    p_operation_id uuid, p_run_id uuid,
    p_seal_receipt_sha256 text, p_seal_receipt_document bytea,
    p_cleanup_receipt_sha256 text, p_cleanup_receipt_document bytea
)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE;
    v_seal jsonb;
    v_cleanup jsonb;
    v_target jsonb;
    v_authority_receipt_sha256 text;
    v_seal_receipt_id uuid;
    v_cleanup_receipt_id uuid;
    v_seal_issued_at timestamptz;
    v_cleanup_issued_at timestamptz;
    v_invalid boolean;
BEGIN
    IF SESSION_USER <> 'harpy_cluster_admin' OR p_operation_id IS NULL OR p_run_id IS NULL OR
       p_seal_receipt_sha256 IS NULL OR p_seal_receipt_sha256!~'^[0-9a-f]{64}$' OR
       p_cleanup_receipt_sha256 IS NULL OR p_cleanup_receipt_sha256!~'^[0-9a-f]{64}$' OR
       p_seal_receipt_document IS NULL OR pg_catalog.octet_length(p_seal_receipt_document) NOT BETWEEN 2 AND 65536 OR
       p_cleanup_receipt_document IS NULL OR pg_catalog.octet_length(p_cleanup_receipt_document) NOT BETWEEN 2 AND 65536 OR
       pg_catalog.encode(pg_catalog.sha256(p_seal_receipt_document),'hex') IS DISTINCT FROM p_seal_receipt_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(p_cleanup_receipt_document),'hex') IS DISTINCT FROM p_cleanup_receipt_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover credential receipt input is invalid';
    END IF;
    SELECT inventory.receipt_sha256 INTO STRICT v_authority_receipt_sha256
    FROM cutover_authority.authority_adoption_inventory AS inventory
    ORDER BY inventory.catalog_generation DESC LIMIT 1;
    PERFORM cutover_authority.assert_adopted_catalog_v1(v_authority_receipt_sha256);
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
        WHERE role.rolname='harpy_cluster_admin' AND role.rolcanlogin AND NOT role.rolinherit
          AND NOT role.rolsuper AND role.rolcreatedb AND role.rolcreaterole AND NOT role.rolreplication
          AND NOT role.rolbypassrls AND role.rolconnlimit=1 AND role.rolvaliduntil>pg_catalog.clock_timestamp()
          AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
              WHERE membership.roleid=role.oid OR membership.member=role.oid)) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover credential recorder role lease is invalid';
    END IF;
    SELECT operation.* INTO STRICT v_operation FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
      AND operation.state IN ('planned','copying','shadow','quiescing','blocked');
    SELECT pg_catalog.jsonb_build_object('database_name',pg_catalog.current_database(),
      'database_origin',release_row.database_origin,'system_identifier',control.system_identifier::text,
      'database_oid',database_row.oid::bigint) INTO STRICT v_target
    FROM platform.release_generations AS release_row CROSS JOIN pg_catalog.pg_control_system() AS control
    CROSS JOIN pg_catalog.pg_database AS database_row WHERE release_row.release_id=v_operation.release_id
      AND release_row.environment_id=v_operation.environment_id AND database_row.datname=pg_catalog.current_database();
    IF v_operation.target_system_identifier::text IS DISTINCT FROM v_target->>'system_identifier' OR
       v_operation.target_database_oid::oid IS DISTINCT FROM (v_target->>'database_oid')::oid THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover credential receipt target differs';
    END IF;
    v_seal:=pg_catalog.convert_from(p_seal_receipt_document,'UTF8')::jsonb;
    v_cleanup:=pg_catalog.convert_from(p_cleanup_receipt_document,'UTF8')::jsonb;
    IF pg_catalog.jsonb_typeof(v_seal) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_seal)) IS DISTINCT FROM 4::bigint OR
       (v_seal ?& ARRAY['contract_version','kind','payload','signature']) IS DISTINCT FROM true OR
       pg_catalog.jsonb_typeof(v_seal->'payload') IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_seal->'payload')) IS DISTINCT FROM 13::bigint OR
       (v_seal->'payload' ?& ARRAY['receipt_id','lease_id','lease_manifest_sha256','environment_id','operation_id','run_id','issued_at','source','target','credentials','source_sessions','target_sessions','result']) IS DISTINCT FROM true OR
       v_seal->>'contract_version' IS DISTINCT FROM 'cutover-session-seal-receipt.v1' OR
       v_seal->>'kind' IS DISTINCT FROM 'cutover_session_seal' OR v_seal->'payload'->>'result' IS DISTINCT FROM 'sealed' OR
       v_seal->'payload'->>'lease_id' IS DISTINCT FROM v_operation.lease_id::text OR
       v_seal->'payload'->>'lease_manifest_sha256' IS DISTINCT FROM v_operation.lease_manifest_sha256 OR
       v_seal->'payload'->>'environment_id' IS DISTINCT FROM v_operation.environment_id::text OR
       v_seal->'payload'->>'operation_id' IS DISTINCT FROM p_operation_id::text OR
       v_seal->'payload'->>'run_id' IS DISTINCT FROM p_run_id::text OR
       v_seal->'payload'->'source' IS DISTINCT FROM v_target OR v_seal->'payload'->'target' IS DISTINCT FROM v_target OR
       v_seal->'payload'->'credentials' IS DISTINCT FROM
         '[{"principal":"harpy_cutover_decoder","can_login":false,"password_state":null,"replication":false,"connection_limit":0},{"principal":"harpy_adopter","can_login":false,"password_state":null,"replication":false,"connection_limit":0}]'::jsonb OR
       pg_catalog.jsonb_typeof(v_seal->'payload'->'source_sessions') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(v_seal->'payload'->'source_sessions') NOT BETWEEN 2 AND 32 OR
       pg_catalog.jsonb_typeof(v_seal->'payload'->'target_sessions') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(v_seal->'payload'->'target_sessions') NOT BETWEEN 1 AND 32 OR
       v_seal->'payload'->>'receipt_id' IS NULL OR
       (v_seal->'payload'->>'receipt_id')::uuid::text IS DISTINCT FROM v_seal->'payload'->>'receipt_id' OR
       v_seal->'payload'->>'issued_at' IS NULL OR v_seal->'payload'->>'issued_at'!~'(Z|[+]00:00)$' OR
       pg_catalog.jsonb_typeof(v_seal->'signature') IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_seal->'signature')) IS DISTINCT FROM 3::bigint OR
       (v_seal->'signature' ?& ARRAY['algorithm','key_id','value']) IS DISTINCT FROM true OR
       v_seal->'signature'->>'algorithm' IS DISTINCT FROM 'Ed25519' OR
       pg_catalog.jsonb_typeof(v_seal->'signature'->'key_id') IS DISTINCT FROM 'string' OR
       pg_catalog.jsonb_typeof(v_seal->'signature'->'value') IS DISTINCT FROM 'string' OR
       v_seal->'signature'->>'key_id'!~'^[a-z0-9][a-z0-9._-]{0,63}$' OR
       v_seal->'signature'->>'value'!~'^[A-Za-z0-9_-]{86}$' THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover session seal receipt differs';
    END IF;
    v_seal_receipt_id:=(v_seal->'payload'->>'receipt_id')::uuid;
    v_seal_issued_at:=(v_seal->'payload'->>'issued_at')::timestamptz;
    IF EXISTS (SELECT 1 FROM pg_catalog.jsonb_array_elements(v_seal->'payload'->'source_sessions') AS session_row
          WHERE (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(session_row)) IS DISTINCT FROM 4::bigint
             OR (session_row ?& ARRAY['pid','backend_start','session_user','application_name']) IS DISTINCT FROM true
             OR pg_catalog.jsonb_typeof(session_row->'pid') IS DISTINCT FROM 'number' OR session_row->>'pid'!~'^[1-9][0-9]*$'
             OR pg_catalog.jsonb_typeof(session_row->'backend_start') IS DISTINCT FROM 'string'
             OR session_row->>'backend_start'!~'(Z|[+]00:00)$' OR (session_row->>'backend_start')::timestamptz>v_seal_issued_at
             OR session_row->>'session_user' IS DISTINCT FROM 'harpy_cutover_decoder'
             OR session_row->>'application_name' IS DISTINCT FROM 'harpy-cutover/'||p_run_id::text) OR
       EXISTS (SELECT 1 FROM pg_catalog.jsonb_array_elements(v_seal->'payload'->'target_sessions') AS session_row
          WHERE (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(session_row)) IS DISTINCT FROM 4::bigint
             OR (session_row ?& ARRAY['pid','backend_start','session_user','application_name']) IS DISTINCT FROM true
             OR pg_catalog.jsonb_typeof(session_row->'pid') IS DISTINCT FROM 'number' OR session_row->>'pid'!~'^[1-9][0-9]*$'
             OR pg_catalog.jsonb_typeof(session_row->'backend_start') IS DISTINCT FROM 'string'
             OR session_row->>'backend_start'!~'(Z|[+]00:00)$' OR (session_row->>'backend_start')::timestamptz>v_seal_issued_at
             OR session_row->>'session_user' IS DISTINCT FROM 'harpy_adopter'
             OR session_row->>'application_name' IS DISTINCT FROM 'harpy-cutover/'||p_run_id::text) THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover sealed session set differs';
    END IF;
    SELECT pg_catalog.count(*) IS DISTINCT FROM pg_catalog.count(DISTINCT session_row.pid) INTO v_invalid
    FROM (SELECT (value->>'pid')::bigint AS pid FROM pg_catalog.jsonb_array_elements(v_seal->'payload'->'source_sessions') AS value
          UNION ALL SELECT (value->>'pid')::bigint FROM pg_catalog.jsonb_array_elements(v_seal->'payload'->'target_sessions') AS value) AS session_row;
    IF v_invalid THEN RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover sealed session pid duplicates'; END IF;
    IF pg_catalog.jsonb_typeof(v_cleanup) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_cleanup)) IS DISTINCT FROM 4::bigint OR
       (v_cleanup ?& ARRAY['contract_version','kind','payload','signature']) IS DISTINCT FROM true OR
       pg_catalog.jsonb_typeof(v_cleanup->'payload') IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_cleanup->'payload')) IS DISTINCT FROM 10::bigint OR
       (v_cleanup->'payload' ?& ARRAY['receipt_id','lease_id','lease_manifest_sha256','seal_receipt_sha256','environment_id','operation_id','run_id','issued_at','principals','result']) IS DISTINCT FROM true OR
       v_cleanup->>'contract_version' IS DISTINCT FROM 'cutover-credential-cleanup-receipt.v1' OR
       v_cleanup->>'kind' IS DISTINCT FROM 'cutover_credential_cleanup' OR v_cleanup->'payload'->>'result' IS DISTINCT FROM 'succeeded' OR
       v_cleanup->'payload'->>'lease_id' IS DISTINCT FROM v_operation.lease_id::text OR
       v_cleanup->'payload'->>'lease_manifest_sha256' IS DISTINCT FROM v_operation.lease_manifest_sha256 OR
       v_cleanup->'payload'->>'seal_receipt_sha256' IS DISTINCT FROM p_seal_receipt_sha256 OR
       v_cleanup->'payload'->>'environment_id' IS DISTINCT FROM v_operation.environment_id::text OR
       v_cleanup->'payload'->>'operation_id' IS DISTINCT FROM p_operation_id::text OR
       v_cleanup->'payload'->>'run_id' IS DISTINCT FROM p_run_id::text OR
       v_cleanup->'payload'->>'receipt_id' IS NULL OR
       (v_cleanup->'payload'->>'receipt_id')::uuid::text IS DISTINCT FROM v_cleanup->'payload'->>'receipt_id' OR
       v_cleanup->'payload'->>'issued_at' IS NULL OR v_cleanup->'payload'->>'issued_at'!~'(Z|[+]00:00)$' OR
       pg_catalog.jsonb_typeof(v_cleanup->'payload'->'principals') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(v_cleanup->'payload'->'principals') IS DISTINCT FROM 2 OR
       ((v_cleanup->'payload'->'principals'->0)-'cleaned_at'::text) IS DISTINCT FROM
         '{"principal":"harpy_cutover_decoder","can_login":false,"password_state":null,"replication":false,"connection_limit":0,"active_sessions":0,"membership_count":0,"dynamic_grants_revoked":true}'::jsonb OR
       ((v_cleanup->'payload'->'principals'->1)-'cleaned_at'::text) IS DISTINCT FROM
         '{"principal":"harpy_adopter","can_login":false,"password_state":null,"replication":false,"connection_limit":0,"active_sessions":0,"membership_count":0,"dynamic_grants_revoked":true}'::jsonb OR
       pg_catalog.jsonb_typeof(v_cleanup->'signature') IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_cleanup->'signature')) IS DISTINCT FROM 3::bigint OR
       (v_cleanup->'signature' ?& ARRAY['algorithm','key_id','value']) IS DISTINCT FROM true OR
       v_cleanup->'signature'->>'algorithm' IS DISTINCT FROM 'Ed25519' OR
       pg_catalog.jsonb_typeof(v_cleanup->'signature'->'key_id') IS DISTINCT FROM 'string' OR
       pg_catalog.jsonb_typeof(v_cleanup->'signature'->'value') IS DISTINCT FROM 'string' OR
       v_cleanup->'signature'->>'key_id'!~'^[a-z0-9][a-z0-9._-]{0,63}$' OR
       v_cleanup->'signature'->>'value'!~'^[A-Za-z0-9_-]{86}$' THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover credential cleanup receipt differs';
    END IF;
    v_cleanup_receipt_id:=(v_cleanup->'payload'->>'receipt_id')::uuid;
    v_cleanup_issued_at:=(v_cleanup->'payload'->>'issued_at')::timestamptz;
    IF v_cleanup_receipt_id=v_seal_receipt_id OR v_cleanup_issued_at<v_seal_issued_at OR
       EXISTS (SELECT 1 FROM pg_catalog.jsonb_array_elements(v_cleanup->'payload'->'principals') AS principal
          WHERE pg_catalog.jsonb_typeof(principal->'cleaned_at') IS DISTINCT FROM 'string'
             OR principal->>'cleaned_at'!~'(Z|[+]00:00)$' OR
             (principal->>'cleaned_at')::timestamptz NOT BETWEEN v_seal_issued_at AND v_cleanup_issued_at) OR
       NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role WHERE role.rolname IN ('harpy_cutover_decoder','harpy_adopter')
          AND NOT role.rolcanlogin AND NOT role.rolinherit AND NOT role.rolsuper AND NOT role.rolcreatedb
          AND NOT role.rolcreaterole AND NOT role.rolreplication AND NOT role.rolbypassrls AND role.rolconnlimit=0
          AND role.rolvaliduntil IS NOT DISTINCT FROM 'infinity'::timestamptz
          AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
             WHERE membership.roleid=role.oid OR membership.member=role.oid)
          HAVING pg_catalog.count(*)=2) OR
       EXISTS (SELECT 1 FROM pg_catalog.pg_stat_activity AS activity
          WHERE activity.usename IN ('harpy_cutover_decoder','harpy_adopter')) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover live credential cleanup state differs';
    END IF;
    INSERT INTO platform.cutover_credential_receipts (receipt_id,operation_id,run_id,lease_id,receipt_kind,
      contract_version,lease_manifest_sha256,seal_receipt_sha256,receipt_sha256,canonical_document,issued_at,verified_at)
    VALUES (v_seal_receipt_id,p_operation_id,p_run_id,v_operation.lease_id,'session_seal',
      'cutover-session-seal-receipt.v1',v_operation.lease_manifest_sha256,NULL,p_seal_receipt_sha256,v_seal,v_seal_issued_at,pg_catalog.clock_timestamp())
    ON CONFLICT DO NOTHING;
    IF NOT FOUND AND NOT EXISTS (SELECT 1 FROM platform.cutover_credential_receipts AS receipt
       WHERE receipt.receipt_id=v_seal_receipt_id AND receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id
         AND receipt.lease_id=v_operation.lease_id AND receipt.receipt_kind='session_seal'
         AND receipt.contract_version='cutover-session-seal-receipt.v1' AND receipt.lease_manifest_sha256=v_operation.lease_manifest_sha256
         AND receipt.seal_receipt_sha256 IS NULL AND receipt.receipt_sha256=p_seal_receipt_sha256
         AND receipt.canonical_document=v_seal AND receipt.issued_at=v_seal_issued_at) THEN
        RAISE EXCEPTION USING ERRCODE='22000', MESSAGE='cutover session seal receipt replay differs';
    END IF;
    INSERT INTO platform.cutover_credential_receipts (receipt_id,operation_id,run_id,lease_id,receipt_kind,
      contract_version,lease_manifest_sha256,seal_receipt_sha256,receipt_sha256,canonical_document,issued_at,verified_at)
    VALUES (v_cleanup_receipt_id,p_operation_id,p_run_id,v_operation.lease_id,'credential_cleanup',
      'cutover-credential-cleanup-receipt.v1',v_operation.lease_manifest_sha256,p_seal_receipt_sha256,p_cleanup_receipt_sha256,v_cleanup,v_cleanup_issued_at,pg_catalog.clock_timestamp())
    ON CONFLICT DO NOTHING;
    IF NOT FOUND AND NOT EXISTS (SELECT 1 FROM platform.cutover_credential_receipts AS receipt
       WHERE receipt.receipt_id=v_cleanup_receipt_id AND receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id
         AND receipt.lease_id=v_operation.lease_id AND receipt.receipt_kind='credential_cleanup'
         AND receipt.contract_version='cutover-credential-cleanup-receipt.v1' AND receipt.lease_manifest_sha256=v_operation.lease_manifest_sha256
         AND receipt.seal_receipt_sha256=p_seal_receipt_sha256 AND receipt.receipt_sha256=p_cleanup_receipt_sha256
         AND receipt.canonical_document=v_cleanup AND receipt.issued_at=v_cleanup_issued_at) THEN
        RAISE EXCEPTION USING ERRCODE='22000', MESSAGE='cutover credential cleanup receipt replay differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.record_credential_receipts_v1(uuid,uuid,text,bytea,text,bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION cutover_authority.record_credential_receipts_v1(uuid,uuid,text,bytea,text,bytea) TO harpy_cluster_admin;
