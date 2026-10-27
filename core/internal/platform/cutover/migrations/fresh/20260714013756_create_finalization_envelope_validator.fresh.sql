-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.validate_finalization_envelope_v1(
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
    v_operation platform.cutover_operations%ROWTYPE;
    v_cluster pg_catalog.pg_roles%ROWTYPE;
    v_finalizer jsonb;
    v_adoption jsonb;
    v_seal jsonb;
    v_cleanup jsonb;
    v_live_relations jsonb;
    v_target jsonb;
    v_state_sha256 text;
    v_live_oid oid; v_system_identifier text;
    v_database_oid oid;
    v_database_origin text;
BEGIN
    IF SESSION_USER NOT IN ('harpy_cluster_admin','harpy_cutover_finalizer') OR
       p_operation_id IS NULL OR p_run_id IS NULL OR p_finalizer_lease_id IS NULL OR
       p_terminal_state NOT IN ('completed','failed','cancelled') OR
       p_finalizer_lease_sha256 IS NULL OR p_finalizer_lease_sha256!~'^[0-9a-f]{64}$' OR
       p_adoption_lease_sha256 IS NULL OR p_adoption_lease_sha256!~'^[0-9a-f]{64}$' OR
       p_seal_receipt_sha256 IS NULL OR p_seal_receipt_sha256!~'^[0-9a-f]{64}$' OR
       p_cleanup_receipt_sha256 IS NULL OR p_cleanup_receipt_sha256!~'^[0-9a-f]{64}$' OR
       p_authority_receipt_sha256 IS NULL OR p_authority_receipt_sha256!~'^[0-9a-f]{64}$' OR
       p_terminal_evidence_sha256 IS NULL OR p_terminal_evidence_sha256!~'^[0-9a-f]{64}$' OR
       p_finalizer_lease_document IS NULL OR pg_catalog.octet_length(p_finalizer_lease_document) NOT BETWEEN 2 AND 65536 OR
       p_adoption_lease_document IS NULL OR pg_catalog.octet_length(p_adoption_lease_document) NOT BETWEEN 2 AND 16777216 OR
       p_seal_receipt_document IS NULL OR pg_catalog.octet_length(p_seal_receipt_document) NOT BETWEEN 2 AND 65536 OR
       p_cleanup_receipt_document IS NULL OR pg_catalog.octet_length(p_cleanup_receipt_document) NOT BETWEEN 2 AND 65536 OR
       p_authority_receipt_document IS NULL OR pg_catalog.octet_length(p_authority_receipt_document) NOT BETWEEN 2 AND 16777216 OR
       p_terminal_evidence_document IS NULL OR pg_catalog.octet_length(p_terminal_evidence_document) NOT BETWEEN 2 AND 16777216 THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover finalization envelope input is invalid';
    END IF;
    IF pg_catalog.encode(pg_catalog.sha256(p_finalizer_lease_document),'hex') IS DISTINCT FROM p_finalizer_lease_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(p_adoption_lease_document),'hex') IS DISTINCT FROM p_adoption_lease_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(p_seal_receipt_document),'hex') IS DISTINCT FROM p_seal_receipt_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(p_cleanup_receipt_document),'hex') IS DISTINCT FROM p_cleanup_receipt_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(p_authority_receipt_document),'hex') IS DISTINCT FROM p_authority_receipt_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(p_terminal_evidence_document),'hex') IS DISTINCT FROM p_terminal_evidence_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='22000', MESSAGE='cutover finalization envelope digest differs';
    END IF;
    v_finalizer:=pg_catalog.convert_from(p_finalizer_lease_document,'UTF8')::jsonb; v_adoption:=pg_catalog.convert_from(p_adoption_lease_document,'UTF8')::jsonb;
    v_seal:=pg_catalog.convert_from(p_seal_receipt_document,'UTF8')::jsonb; v_cleanup:=pg_catalog.convert_from(p_cleanup_receipt_document,'UTF8')::jsonb;
    v_state_sha256:=cutover_authority.finalization_state_sha256_v1(p_operation_id,p_run_id);
    SELECT operation.* INTO STRICT v_operation FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id;
    SELECT role.* INTO STRICT v_cluster FROM pg_catalog.pg_roles AS role
    WHERE role.rolname='harpy_cluster_admin' AND role.rolcanlogin AND NOT role.rolinherit
      AND NOT role.rolsuper AND role.rolcreatedb AND role.rolcreaterole AND NOT role.rolreplication
      AND NOT role.rolbypassrls AND role.rolconnlimit=1 AND role.rolvaliduntil>pg_catalog.clock_timestamp()
      AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
          WHERE membership.roleid=role.oid OR membership.member=role.oid);
    SELECT release_row.database_origin,control.system_identifier::text,database_row.oid
    INTO STRICT v_database_origin,v_system_identifier,v_database_oid
    FROM platform.release_generations AS release_row CROSS JOIN pg_catalog.pg_control_system() AS control
    CROSS JOIN pg_catalog.pg_database AS database_row
    WHERE release_row.release_id=v_operation.release_id AND release_row.environment_id=v_operation.environment_id
      AND database_row.datname=pg_catalog.current_database();
    v_target:=pg_catalog.jsonb_build_object('database_name',pg_catalog.current_database(),
        'database_origin',v_database_origin,'system_identifier',v_system_identifier,
        'database_oid',v_database_oid::bigint);
    v_live_oid:='cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)'::regprocedure::oid;
    IF pg_catalog.jsonb_typeof(v_finalizer) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_finalizer)) IS DISTINCT FROM 22::bigint OR
       (v_finalizer ?& ARRAY['contract_version','lease_id','release_id','environment_id','operation_id','run_id','issued_at','not_before','expires_at','max_duration_seconds','role_lease','target','adoption_lease_sha256','cleanup_receipt_sha256','authority_adoption_receipt_sha256','finalization_state_sha256','tenant_requirement_catalog_sha256','tenant_requirement_count','finalization_procedure_oid','finalization_procedure_signature','terminal_state','terminal_evidence_sha256']) IS DISTINCT FROM true OR
       v_finalizer->>'contract_version' IS DISTINCT FROM 'cutover-finalizer-lease.v1' OR
       (v_finalizer->>'lease_id')::uuid IS DISTINCT FROM p_finalizer_lease_id OR
       (v_finalizer->>'operation_id')::uuid IS DISTINCT FROM p_operation_id OR
       (v_finalizer->>'run_id')::uuid IS DISTINCT FROM p_run_id OR
       (v_finalizer->>'release_id')::uuid IS DISTINCT FROM v_operation.release_id OR
       (v_finalizer->>'environment_id')::uuid IS DISTINCT FROM v_operation.environment_id OR
       v_finalizer->>'adoption_lease_sha256' IS DISTINCT FROM p_adoption_lease_sha256 OR
       v_finalizer->>'cleanup_receipt_sha256' IS DISTINCT FROM p_cleanup_receipt_sha256 OR
       v_finalizer->>'authority_adoption_receipt_sha256' IS DISTINCT FROM p_authority_receipt_sha256 OR
       v_finalizer->>'finalization_state_sha256' IS DISTINCT FROM v_state_sha256 OR
       v_finalizer->>'terminal_state' IS DISTINCT FROM p_terminal_state OR
       v_finalizer->>'terminal_evidence_sha256' IS DISTINCT FROM p_terminal_evidence_sha256 OR
       v_finalizer->'role_lease' IS DISTINCT FROM '{"principal":"harpy_cutover_finalizer","login":true,"replication":false,"connection_limit":1,"password_state":"leased_secret"}'::jsonb OR
       v_finalizer->'target' IS DISTINCT FROM v_target OR
       (v_finalizer->>'finalization_procedure_oid')::oid IS DISTINCT FROM v_live_oid OR
       v_finalizer->>'finalization_procedure_signature' IS DISTINCT FROM 'cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)' OR
       v_finalizer->>'max_duration_seconds' IS DISTINCT FROM '120' OR
       (v_finalizer->>'issued_at')::timestamptz>(v_finalizer->>'not_before')::timestamptz OR
       (v_finalizer->>'not_before')::timestamptz>pg_catalog.clock_timestamp() OR
       (v_finalizer->>'expires_at')::timestamptz<=pg_catalog.clock_timestamp() OR
       (v_finalizer->>'expires_at')::timestamptz-(v_finalizer->>'issued_at')::timestamptz>interval '120 seconds' OR
       (v_finalizer->>'expires_at')::timestamptz-(v_finalizer->>'not_before')::timestamptz>interval '120 seconds' OR
       (v_finalizer->>'expires_at')::timestamptz>v_cluster.rolvaliduntil OR
       NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
          WHERE role.rolname='harpy_cutover_finalizer' AND role.rolcanlogin AND NOT role.rolinherit
            AND NOT role.rolsuper AND NOT role.rolcreatedb AND NOT role.rolcreaterole
            AND NOT role.rolreplication AND NOT role.rolbypassrls AND role.rolconnlimit=1
            AND role.rolvaliduntil IS NOT DISTINCT FROM (v_finalizer->>'expires_at')::timestamptz
            AND role.rolvaliduntil>pg_catalog.clock_timestamp() AND NOT EXISTS
              (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
               WHERE membership.roleid=role.oid OR membership.member=role.oid)) THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover finalizer lease differs';
    END IF;
    IF (p_terminal_state='completed' AND NOT EXISTS (SELECT 1
        FROM platform.release_generations AS release_row
        JOIN platform.tenant_migration_requirement_catalogs AS catalog
          ON catalog.migration_run_id=release_row.migration_run_id AND catalog.release_id=release_row.release_id
         AND catalog.environment_id=release_row.environment_id AND catalog.database_origin=release_row.database_origin
        WHERE release_row.release_id=v_operation.release_id AND release_row.environment_id=v_operation.environment_id
          AND catalog.catalog_sha256=v_finalizer->>'tenant_requirement_catalog_sha256'
          AND catalog.requirement_count=(v_finalizer->>'tenant_requirement_count')::integer)) OR
       (p_terminal_state IN ('failed','cancelled') AND
        (v_finalizer->'tenant_requirement_catalog_sha256' IS DISTINCT FROM 'null'::jsonb OR
         v_finalizer->'tenant_requirement_count' IS DISTINCT FROM 'null'::jsonb)) THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover tenant authority lease differs';
    END IF;
    SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object('ordinal',relation.ordinal,
        'source',pg_catalog.jsonb_build_object('schema_oid',relation.source_schema_oid,'relation_oid',relation.source_relation_oid,'schema_name',relation.source_schema_name,'relation_name',relation.source_relation_name),
        'target',pg_catalog.jsonb_build_object('schema_oid',relation.target_schema_oid,'relation_oid',relation.target_relation_oid,'schema_name',relation.target_schema_name,'relation_name',relation.target_relation_name),
        'replica_identity',relation.replica_identity,'columns',(SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object('ordinal',column_row.ordinal,'source_attnum',column_row.source_attnum,'target_attnum',column_row.target_attnum,'transform_id',column_row.transform_id,'primary_key',column_row.primary_key) ORDER BY column_row.ordinal) FROM platform.cutover_operation_columns AS column_row WHERE column_row.operation_id=p_operation_id AND column_row.run_id=p_run_id AND column_row.relation_ordinal=relation.ordinal),
        'procedure_contract_id',procedure_row.procedure_contract_id,'apply_procedure_oid',procedure_row.procedure_oid,'apply_procedure_signature',procedure_row.procedure_signature,
        'expected_checkpoint_keys',(SELECT pg_catalog.jsonb_agg(checkpoint_key.checkpoint_key ORDER BY checkpoint_key.ordinal) FROM platform.cutover_operation_checkpoint_keys AS checkpoint_key WHERE checkpoint_key.operation_id=p_operation_id AND checkpoint_key.run_id=p_run_id AND checkpoint_key.relation_ordinal=relation.ordinal)) ORDER BY relation.ordinal),'[]'::jsonb)
    INTO v_live_relations FROM platform.cutover_operation_relations AS relation
    JOIN platform.cutover_operation_procedures AS procedure_row ON procedure_row.operation_id=relation.operation_id AND procedure_row.run_id=relation.run_id AND procedure_row.relation_ordinal=relation.ordinal AND procedure_row.procedure_kind='apply'
    WHERE relation.operation_id=p_operation_id AND relation.run_id=p_run_id;
    IF pg_catalog.jsonb_typeof(v_adoption) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_adoption)) IS DISTINCT FROM 24::bigint OR
       (v_adoption ?& ARRAY['contract_version','lease_id','release_id','environment_id','operation_id','run_id','issued_at','not_before','expires_at','max_duration_seconds','scope_kind','scope_id','group_id','domains','plan_sha256','transform_manifest_sha256','source','target','publication_name','slot_name','source_connection_limit','target_connection_limit','role_leases','relations']) IS DISTINCT FROM true OR
       v_adoption->>'contract_version' IS DISTINCT FROM 'cutover-credential-lease.v1' OR
       (v_adoption->>'lease_id')::uuid IS DISTINCT FROM v_operation.lease_id OR
       (v_adoption->>'operation_id')::uuid IS DISTINCT FROM p_operation_id OR
       (v_adoption->>'run_id')::uuid IS DISTINCT FROM p_run_id OR
       (v_adoption->>'release_id')::uuid IS DISTINCT FROM v_operation.release_id OR
       (v_adoption->>'environment_id')::uuid IS DISTINCT FROM v_operation.environment_id OR
       v_adoption->>'scope_kind' IS DISTINCT FROM v_operation.scope_kind OR
       (v_adoption->>'scope_id')::uuid IS DISTINCT FROM v_operation.scope_id OR
       v_adoption->>'group_id' IS DISTINCT FROM v_operation.group_id OR
       (v_adoption->'domains' ? v_operation.domain) IS DISTINCT FROM true OR
       v_adoption->>'plan_sha256' IS DISTINCT FROM v_operation.plan_sha256 OR
       v_adoption->>'transform_manifest_sha256' IS DISTINCT FROM v_operation.transform_manifest_sha256 OR
       v_adoption->'source' IS DISTINCT FROM v_target OR v_adoption->'target' IS DISTINCT FROM v_target OR
       v_adoption->>'publication_name' IS DISTINCT FROM v_operation.publication_name OR
       v_adoption->>'slot_name' IS DISTINCT FROM v_operation.slot_name OR
       (v_adoption->>'expires_at')::timestamptz IS DISTINCT FROM v_operation.lease_expires_at OR
       v_adoption->'relations' IS DISTINCT FROM v_live_relations OR
       p_adoption_lease_sha256 IS DISTINCT FROM v_operation.lease_manifest_sha256 OR
       EXISTS (SELECT 1 FROM platform.cutover_operation_checkpoint_keys AS checkpoint_key
          WHERE checkpoint_key.operation_id=p_operation_id AND checkpoint_key.run_id=p_run_id
          GROUP BY checkpoint_key.relation_ordinal HAVING pg_catalog.min(checkpoint_key.ordinal)<>1
            OR pg_catalog.max(checkpoint_key.ordinal)<>pg_catalog.count(*)
            OR pg_catalog.count(DISTINCT checkpoint_key.ordinal)<>pg_catalog.count(*)) OR
       (SELECT pg_catalog.count(*) FROM platform.cutover_operation_checkpoint_keys AS checkpoint_key
          WHERE checkpoint_key.operation_id=p_operation_id AND checkpoint_key.run_id=p_run_id)>10000 THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover adoption projection differs';
    END IF;
    IF (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_seal)) IS DISTINCT FROM 4::bigint OR
       (v_seal ?& ARRAY['contract_version','kind','payload','signature']) IS DISTINCT FROM true OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_seal->'payload')) IS DISTINCT FROM 13::bigint OR
       (v_seal->'payload' ?& ARRAY['receipt_id','lease_id','lease_manifest_sha256','environment_id','operation_id','run_id','issued_at','source','target','credentials','source_sessions','target_sessions','result']) IS DISTINCT FROM true OR
       v_seal->>'contract_version' IS DISTINCT FROM 'cutover-session-seal-receipt.v1' OR v_seal->>'kind' IS DISTINCT FROM 'cutover_session_seal' OR
       v_seal->'payload'->>'lease_id' IS DISTINCT FROM v_operation.lease_id::text OR v_seal->'payload'->>'operation_id' IS DISTINCT FROM p_operation_id::text OR
       v_seal->'payload'->>'run_id' IS DISTINCT FROM p_run_id::text OR v_seal->'payload'->>'environment_id' IS DISTINCT FROM v_operation.environment_id::text OR
       v_seal->'payload'->>'lease_manifest_sha256' IS DISTINCT FROM p_adoption_lease_sha256 OR v_seal->'payload'->'source' IS DISTINCT FROM v_target OR
       v_seal->'payload'->'target' IS DISTINCT FROM v_target OR v_seal->'payload'->>'result' IS DISTINCT FROM 'sealed' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_cleanup)) IS DISTINCT FROM 4::bigint OR
       (v_cleanup ?& ARRAY['contract_version','kind','payload','signature']) IS DISTINCT FROM true OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_cleanup->'payload')) IS DISTINCT FROM 10::bigint OR
       (v_cleanup->'payload' ?& ARRAY['receipt_id','lease_id','lease_manifest_sha256','seal_receipt_sha256','environment_id','operation_id','run_id','issued_at','principals','result']) IS DISTINCT FROM true OR
       v_cleanup->>'contract_version' IS DISTINCT FROM 'cutover-credential-cleanup-receipt.v1' OR v_cleanup->>'kind' IS DISTINCT FROM 'cutover_credential_cleanup' OR
       v_cleanup->'payload'->>'lease_id' IS DISTINCT FROM v_operation.lease_id::text OR v_cleanup->'payload'->>'operation_id' IS DISTINCT FROM p_operation_id::text OR
       v_cleanup->'payload'->>'run_id' IS DISTINCT FROM p_run_id::text OR v_cleanup->'payload'->>'environment_id' IS DISTINCT FROM v_operation.environment_id::text OR
       v_cleanup->'payload'->>'lease_manifest_sha256' IS DISTINCT FROM p_adoption_lease_sha256 OR v_cleanup->'payload'->>'seal_receipt_sha256' IS DISTINCT FROM p_seal_receipt_sha256 OR
       v_cleanup->'payload'->>'result' IS DISTINCT FROM 'succeeded' OR
       NOT EXISTS (SELECT 1 FROM platform.cutover_credential_receipts AS receipt WHERE receipt.receipt_id=(v_seal->'payload'->>'receipt_id')::uuid AND receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id AND receipt.lease_id=v_operation.lease_id AND receipt.receipt_kind='session_seal' AND receipt.lease_manifest_sha256=p_adoption_lease_sha256 AND receipt.receipt_sha256=p_seal_receipt_sha256 AND receipt.canonical_document IS NOT DISTINCT FROM v_seal) OR
       NOT EXISTS (SELECT 1 FROM platform.cutover_credential_receipts AS receipt WHERE receipt.receipt_id=(v_cleanup->'payload'->>'receipt_id')::uuid AND receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id AND receipt.lease_id=v_operation.lease_id AND receipt.receipt_kind='credential_cleanup' AND receipt.lease_manifest_sha256=p_adoption_lease_sha256 AND receipt.seal_receipt_sha256=p_seal_receipt_sha256 AND receipt.receipt_sha256=p_cleanup_receipt_sha256 AND receipt.canonical_document IS NOT DISTINCT FROM v_cleanup) THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover cleanup evidence differs';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role WHERE role.rolname IN ('harpy_cutover_decoder','harpy_adopter')
        AND NOT role.rolcanlogin AND NOT role.rolinherit AND NOT role.rolsuper AND NOT role.rolcreatedb
        AND NOT role.rolcreaterole AND NOT role.rolreplication AND NOT role.rolbypassrls AND role.rolconnlimit=0
        AND role.rolvaliduntil IS NOT DISTINCT FROM 'infinity'::timestamptz
        AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership WHERE membership.roleid=role.oid OR membership.member=role.oid)
        HAVING pg_catalog.count(*)=2) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='cutover cleanup role state differs';
    END IF;
    PERFORM cutover_authority.assert_adopted_catalog_v1(p_authority_receipt_sha256);
    IF NOT EXISTS (SELECT 1 FROM cutover_authority.authority_adoption_inventory AS inventory
        WHERE inventory.receipt_sha256=p_authority_receipt_sha256 AND inventory.receipt_document=p_authority_receipt_document) THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='cutover authority receipt bytes differ';
    END IF;
    PERFORM cutover_authority.assert_terminal_evidence_v1(p_operation_id,p_run_id,p_terminal_state,p_terminal_evidence_document);
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.validate_finalization_envelope_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea) FROM PUBLIC;
