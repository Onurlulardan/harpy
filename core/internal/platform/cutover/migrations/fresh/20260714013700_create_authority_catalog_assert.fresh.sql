-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.assert_adopted_catalog_v1(p_receipt_sha256 text)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_inventory cutover_authority.authority_adoption_inventory%ROWTYPE;
    v_receipt jsonb; v_item jsonb;
    v_role_evidence jsonb;
    v_at_rest jsonb;
    v_temporary jsonb;
    v_oid oid;
    v_owner text;
    v_definition text; v_acl text;
    v_count bigint;
    v_system_identifier text;
    v_database_oid oid;
    v_protected boolean;
BEGIN
    IF p_receipt_sha256 IS NULL OR p_receipt_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='authority receipt identity is invalid';
    END IF;
    SELECT inventory.* INTO STRICT v_inventory FROM cutover_authority.authority_adoption_inventory AS inventory
    ORDER BY inventory.catalog_generation DESC LIMIT 1;
    IF v_inventory.receipt_sha256 IS DISTINCT FROM p_receipt_sha256 OR
       pg_catalog.encode(pg_catalog.sha256(v_inventory.receipt_document),'hex') IS DISTINCT FROM p_receipt_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='current authority receipt digest differs';
    END IF;
    v_receipt := pg_catalog.convert_from(v_inventory.receipt_document,'UTF8')::jsonb;
    IF pg_catalog.jsonb_typeof(v_receipt) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_receipt)) IS DISTINCT FROM 32::bigint OR
       pg_catalog.jsonb_typeof(v_receipt->'objects') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_typeof(v_receipt->'dependencies') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_typeof(v_receipt->'target') IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_receipt->'target')) IS DISTINCT FROM 4::bigint OR
       v_receipt->>'contract_version' IS DISTINCT FROM 'cutover-authority-adoption-receipt.v1' OR
       v_receipt->>'result' IS DISTINCT FROM 'succeeded' OR
       pg_catalog.jsonb_typeof(v_receipt->'catalog_generation') IS DISTINCT FROM 'number' OR
       v_receipt->>'catalog_generation' IS NULL OR
       NOT (v_receipt->>'catalog_generation' ~ '^[1-9][0-9]*$') OR
       (v_receipt->>'catalog_generation')::integer IS DISTINCT FROM v_inventory.catalog_generation OR
       v_receipt->>'receipt_id' IS DISTINCT FROM v_inventory.receipt_id::text OR
       v_receipt->>'operation_id' IS DISTINCT FROM v_inventory.operation_id::text OR
       v_receipt->>'operation_sha256' IS DISTINCT FROM v_inventory.operation_sha256 OR
       v_receipt->>'release_id' IS DISTINCT FROM v_inventory.release_id::text OR
       v_receipt->>'environment_id' IS DISTINCT FROM v_inventory.environment_id::text OR
       v_receipt->'target'->>'database_origin' IS DISTINCT FROM v_inventory.database_origin OR
       v_receipt->>'post_catalog_contract_sha256' IS DISTINCT FROM v_inventory.catalog_contract_sha256 OR
       v_receipt->>'catalog_instance_sha256' IS DISTINCT FROM v_inventory.catalog_instance_sha256 OR
       v_receipt->>'dependency_contract_sha256' IS DISTINCT FROM v_inventory.dependency_contract_sha256 OR
       v_receipt->>'dependency_instance_sha256' IS DISTINCT FROM v_inventory.dependency_instance_sha256 OR
       pg_catalog.jsonb_array_length(v_receipt->'objects') IS DISTINCT FROM v_inventory.object_count OR
       pg_catalog.jsonb_array_length(v_receipt->'dependencies') IS DISTINCT FROM v_inventory.dependency_count THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='current authority receipt projection differs';
    END IF;
    SELECT control.system_identifier::text,database_row.oid INTO STRICT v_system_identifier,v_database_oid
    FROM pg_catalog.pg_control_system() AS control CROSS JOIN pg_catalog.pg_database AS database_row
    WHERE database_row.datname=pg_catalog.current_database();
    IF v_receipt->'target'->>'database_name' IS DISTINCT FROM pg_catalog.current_database() OR
       v_receipt->'target'->>'system_identifier' IS DISTINCT FROM v_system_identifier OR
       v_receipt->'target'->>'database_oid' IS NULL OR
       NOT (v_receipt->'target'->>'database_oid' ~ '^[1-9][0-9]*$') OR
       (v_receipt->'target'->>'database_oid')::oid IS DISTINCT FROM v_database_oid OR
       v_inventory.system_identifier::text IS DISTINCT FROM v_system_identifier OR
       v_inventory.database_oid IS DISTINCT FROM v_database_oid THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='authority database identity differs';
    END IF;
    IF pg_catalog.jsonb_typeof(v_receipt->'quiescence_evidence') IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_receipt->'quiescence_evidence')) IS DISTINCT FROM 4::bigint OR
       pg_catalog.jsonb_typeof(v_receipt->'quiescence_evidence'->'roles') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(v_receipt->'quiescence_evidence'->'roles') IS DISTINCT FROM 1 OR
       v_receipt->'quiescence_evidence'->>'restored_before_commit' IS DISTINCT FROM 'true' OR
       v_receipt->'quiescence_evidence'->>'observed_pre_namespace_sha256' IS NULL OR NOT (v_receipt->'quiescence_evidence'->>'observed_pre_namespace_sha256' ~ '^[0-9a-f]{64}$') OR
       v_receipt->'quiescence_evidence'->>'observed_post_namespace_sha256' IS NULL OR NOT (v_receipt->'quiescence_evidence'->>'observed_post_namespace_sha256' ~ '^[0-9a-f]{64}$') THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='authority quiescence evidence differs';
    END IF;
    v_role_evidence := v_receipt->'quiescence_evidence'->'roles'->0;
    v_at_rest := v_role_evidence->'observed_pre_state';
    v_temporary := v_at_rest || '{"login":false,"connection_limit":0}'::jsonb;
    IF pg_catalog.jsonb_typeof(v_role_evidence) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_role_evidence)) IS DISTINCT FROM 7::bigint OR
       v_role_evidence->>'principal' IS DISTINCT FROM 'forgecrud_migrator' OR
       (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_at_rest)) IS DISTINCT FROM 11::bigint OR
       (v_at_rest-'password_state'::text) IS DISTINCT FROM '{"login":true,"inherit":false,"superuser":false,"createdb":false,"createrole":false,"replication":false,"bypassrls":false,"connection_limit":-1,"valid_until":"infinity","membership_count":1}'::jsonb OR
       v_at_rest->>'password_state' IS NULL OR v_at_rest->>'password_state' NOT IN ('null','scram_sha_256') OR
       v_role_evidence->'observed_temporary_state' IS DISTINCT FROM v_temporary OR
       v_role_evidence->'observed_post_state' IS DISTINCT FROM v_at_rest OR
       pg_catalog.jsonb_typeof(v_role_evidence->'terminated_session_count') IS DISTINCT FROM 'number' OR
       v_role_evidence->>'terminated_session_count' !~ '^[0-9]+$' OR
       (v_role_evidence->>'terminated_session_count')::numeric > 1000000 OR
       v_role_evidence->>'remaining_session_count' IS DISTINCT FROM '0' OR
       v_role_evidence->>'password_verifier_unchanged' IS DISTINCT FROM 'true' THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='authority quiesced role evidence differs';
    END IF;
    IF ((v_receipt->'authority_owner_role')-'owned_object_count'::text) IS DISTINCT FROM '{"login":false,"inherit":false,"superuser":false,"createdb":false,"createrole":false,"replication":false,"bypassrls":false,"connection_limit":-1,"password_state":"null","valid_until":"infinity","membership_count":0}'::jsonb OR
       v_receipt->'authority_owner_role'->>'owned_object_count' IS NULL OR NOT (v_receipt->'authority_owner_role'->>'owned_object_count' ~ '^[1-9][0-9]*$') OR
       (v_receipt->'authority_owner_role'->>'owned_object_count')::integer IS DISTINCT FROM v_inventory.object_count THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='authority owner receipt state differs';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
        JOIN pg_catalog.pg_auth_members AS membership ON membership.member=role.oid
        JOIN pg_catalog.pg_roles AS granted ON granted.oid=membership.roleid
        JOIN pg_catalog.pg_roles AS grantor ON grantor.oid=membership.grantor
        WHERE role.rolname='forgecrud_migrator' AND role.rolcanlogin AND NOT role.rolinherit
          AND NOT role.rolsuper AND NOT role.rolcreatedb AND NOT role.rolcreaterole
          AND NOT role.rolreplication AND NOT role.rolbypassrls AND role.rolconnlimit=-1
          AND (role.rolvaliduntil IS NULL OR
               role.rolvaliduntil='infinity'::timestamptz)
          AND granted.rolname='forgecrud_owner'
          AND grantor.rolname=v_receipt->>'executor_session_user'
          AND NOT membership.admin_option AND NOT membership.inherit_option AND membership.set_option
          AND (SELECT pg_catalog.count(*) FROM pg_catalog.pg_auth_members AS all_membership
               WHERE all_membership.member=role.oid) = 1
          AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS child WHERE child.roleid=role.oid)) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='authority migrator post-state differs';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
        WHERE role.rolname='forgecrud_cutover_authority_owner' AND NOT role.rolcanlogin
          AND NOT role.rolinherit AND NOT role.rolsuper AND NOT role.rolcreatedb
          AND NOT role.rolcreaterole AND NOT role.rolreplication AND NOT role.rolbypassrls
          AND role.rolconnlimit=-1 AND role.rolvaliduntil IS NOT DISTINCT FROM 'infinity'::timestamptz
          AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
              WHERE membership.roleid=role.oid OR membership.member=role.oid)) THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='authority owner state differs';
    END IF;
    SELECT pg_catalog.count(DISTINCT pg_catalog.jsonb_build_array(item->>'kind',item->>'schema_name',
        item->>'object_name',item->>'identity_arguments',item->>'object_oid')) INTO v_count
    FROM pg_catalog.jsonb_array_elements(v_receipt->'objects') AS item;
    IF v_count IS DISTINCT FROM v_inventory.object_count::bigint THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='authority object inventory duplicates';
    END IF;
    SELECT pg_catalog.count(DISTINCT pg_catalog.jsonb_build_array(item->>'kind',item->>'schema_name',
        item->>'object_name',item->>'identity_arguments',item->>'object_oid')) INTO v_count
    FROM pg_catalog.jsonb_array_elements(v_receipt->'dependencies') AS item;
    IF v_count IS DISTINCT FROM v_inventory.dependency_count::bigint THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='authority dependency inventory duplicates';
    END IF;
    FOR v_item,v_protected IN
        SELECT value,true FROM pg_catalog.jsonb_array_elements(v_receipt->'objects')
        UNION ALL SELECT value,false FROM pg_catalog.jsonb_array_elements(v_receipt->'dependencies')
    LOOP
        IF pg_catalog.jsonb_typeof(v_item) IS DISTINCT FROM 'object' OR
           (SELECT pg_catalog.count(*) FROM pg_catalog.jsonb_object_keys(v_item))
             IS DISTINCT FROM (CASE WHEN v_protected THEN 9::bigint ELSE 8::bigint END) OR
           pg_catalog.jsonb_typeof(v_item->'object_oid') IS DISTINCT FROM 'number' OR
           v_item->>'object_oid' IS NULL OR NOT (v_item->>'object_oid' ~ '^[1-9][0-9]*$') OR
           pg_catalog.jsonb_typeof(v_item->'kind') IS DISTINCT FROM 'string' OR
           pg_catalog.jsonb_typeof(v_item->'schema_name') IS DISTINCT FROM 'string' OR
           pg_catalog.jsonb_typeof(v_item->'object_name') IS DISTINCT FROM 'string' OR
           pg_catalog.jsonb_typeof(v_item->'identity_arguments') IS DISTINCT FROM 'string' OR
           pg_catalog.jsonb_typeof(v_item->'owner') IS DISTINCT FROM 'string' OR
           pg_catalog.jsonb_typeof(v_item->'definition_sha256') IS DISTINCT FROM 'string' OR
           pg_catalog.jsonb_typeof(v_item->'acl_sha256') IS DISTINCT FROM 'string' OR
           v_item->>'kind' NOT IN ('schema','table','function') OR
           (NOT v_protected AND v_item->>'kind'='function') OR
           (v_item->>'kind' IN ('schema','table') AND v_item->>'identity_arguments' IS DISTINCT FROM '') OR
           (v_item->>'kind'='schema' AND v_item->>'object_name' IS DISTINCT FROM v_item->>'schema_name') THEN
            RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='authority catalog item shape differs';
        END IF;
        v_oid := (v_item->>'object_oid')::oid;
        SELECT pg_catalog.count(*),pg_catalog.min(located.owner) INTO v_count,v_owner FROM (
            SELECT namespace.oid,pg_catalog.pg_get_userbyid(namespace.nspowner) AS owner
            FROM pg_catalog.pg_namespace AS namespace WHERE v_item->>'kind'='schema'
              AND namespace.nspname=v_item->>'schema_name' AND namespace.oid=v_oid
            UNION ALL SELECT relation.oid,pg_catalog.pg_get_userbyid(relation.relowner)
            FROM pg_catalog.pg_class AS relation JOIN pg_catalog.pg_namespace AS namespace
              ON namespace.oid=relation.relnamespace WHERE v_item->>'kind'='table'
              AND namespace.nspname=v_item->>'schema_name' AND relation.relname=v_item->>'object_name'
              AND relation.relkind IN ('r','p') AND relation.oid=v_oid
            UNION ALL SELECT procedure_row.oid,pg_catalog.pg_get_userbyid(procedure_row.proowner)
            FROM pg_catalog.pg_proc AS procedure_row JOIN pg_catalog.pg_namespace AS namespace
              ON namespace.oid=procedure_row.pronamespace WHERE v_item->>'kind'='function'
              AND namespace.nspname=v_item->>'schema_name' AND procedure_row.proname=v_item->>'object_name'
              AND pg_catalog.replace(pg_catalog.oidvectortypes(procedure_row.proargtypes),' ','')=v_item->>'identity_arguments'
              AND procedure_row.oid=v_oid
        ) AS located;
        SELECT cutover_authority.catalog_definition_sha256_v1(v_item->>'kind',v_oid),
               cutover_authority.catalog_acl_sha256_v1(v_item->>'kind',v_oid)
        INTO v_definition,v_acl;
        IF v_count IS DISTINCT FROM 1 OR v_owner IS DISTINCT FROM v_item->>'owner' OR
           v_definition IS DISTINCT FROM v_item->>'definition_sha256' OR
           v_acl IS DISTINCT FROM v_item->>'acl_sha256' THEN
            RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='authority catalog fingerprint differs';
        END IF;
    END LOOP;
    SELECT pg_catalog.count(*) INTO v_count FROM (
        SELECT namespace.oid FROM pg_catalog.pg_namespace AS namespace
        WHERE namespace.nspname='cutover_authority' AND pg_catalog.pg_get_userbyid(namespace.nspowner)='forgecrud_cutover_authority_owner'
        UNION ALL SELECT relation.oid FROM pg_catalog.pg_class AS relation JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
        WHERE namespace.nspname='cutover_authority' AND relation.relkind IN ('r','p') AND pg_catalog.pg_get_userbyid(relation.relowner)='forgecrud_cutover_authority_owner'
        UNION ALL SELECT procedure_row.oid FROM pg_catalog.pg_proc AS procedure_row JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure_row.pronamespace
        WHERE namespace.nspname='cutover_authority' AND pg_catalog.pg_get_userbyid(procedure_row.proowner)='forgecrud_cutover_authority_owner'
    ) AS live_object;
    IF v_count IS DISTINCT FROM v_inventory.object_count::bigint THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='authority live object set differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.assert_adopted_catalog_v1(text) FROM PUBLIC;
