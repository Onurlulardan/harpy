-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION audit_v1_adopt_batch_v1(
    p_operation_id uuid, p_run_id uuid, p_limit integer, p_observed_at timestamptz
) RETURNS jsonb
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE
    v_run audit_v1_adoption_runs%ROWTYPE;
    v_existing audit_v1_adoption_runs%ROWTYPE;
    v_receipt audit_v1_adoption_receipts%ROWTYPE;
    v_relation regclass;
    v_source record;
    v_source_document jsonb;
    v_changes jsonb;
    v_source_sha bytea;
    v_target_sha bytea;
    v_request_sha bytea;
    v_source_hex text := '';
    v_target_hex text := '';
    v_target_id uuid;
    v_owner_id uuid;
    v_actor_id uuid;
    v_org_id uuid;
    v_correlation_id uuid;
    v_owner_module text;
    v_entity_type text;
    v_form_slug text;
    v_operation text;
    v_occurred_at timestamptz;
    v_cursor_at timestamp without time zone;
    v_cursor_id uuid;
    v_source_count bigint;
    v_drift_count bigint;
    v_rows integer := 0;
    v_page bigint;
    v_response jsonb;
BEGIN
    IF pg_catalog.substring(p_operation_id::text,15,1) <> '7'
       OR pg_catalog.substring(p_run_id::text,15,1) <> '7'
       OR p_limit NOT BETWEEN 1 AND 100 OR p_observed_at IS NULL
       OR pg_catalog.date_trunc('microseconds',p_observed_at) <> p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='AUDIT_V1_BATCH_INVALID';
    END IF;
    v_request_sha := pg_catalog.sha256(pg_catalog.uuid_send(p_run_id) ||
        pg_catalog.int4send(p_limit));
    SELECT receipt.* INTO v_receipt FROM audit_v1_adoption_receipts AS receipt
     WHERE receipt.operation_id=p_operation_id;
    IF FOUND THEN
        IF v_receipt.run_id <> p_run_id OR v_receipt.request_sha256 <> v_request_sha THEN
            RAISE EXCEPTION USING ERRCODE='P6802', MESSAGE='AUDIT_V1_OPERATION_REUSED';
        END IF;
        RETURN v_receipt.response_document;
    END IF;

    v_relation := pg_catalog.to_regclass('change_logs');
    SELECT run.* INTO v_existing FROM audit_v1_adoption_runs AS run
     WHERE run.singleton FOR UPDATE;
    IF FOUND AND v_existing.run_id <> p_run_id THEN
        RAISE EXCEPTION USING ERRCODE='P6802', MESSAGE='AUDIT_V1_RUN_ALREADY_EXISTS';
    END IF;
    IF NOT FOUND THEN
        IF v_relation IS NULL THEN
            v_source_count := 0;
        ELSE
            EXECUTE pg_catalog.format('SELECT count(*) FROM %s',v_relation)
               INTO v_source_count;
        END IF;
        INSERT INTO audit_v1_adoption_runs (
            run_id,source_relation_oid,status,source_count,adopted_count,next_page,
            cursor_created_at,cursor_id,started_at
        ) VALUES (
            p_run_id,v_relation::oid,'applying',v_source_count,0,1,
            '-infinity'::timestamp,'00000000-0000-0000-0000-000000000000',p_observed_at
        ) RETURNING * INTO v_run;
    ELSE
        v_run := v_existing;
    END IF;
    IF v_run.source_relation_oid IS DISTINCT FROM v_relation::oid THEN
        RAISE EXCEPTION USING ERRCODE='P6803', MESSAGE='AUDIT_V1_SOURCE_RELATION_DRIFT';
    END IF;
    v_page := v_run.next_page;

    IF v_run.status='applying' AND v_relation IS NOT NULL THEN
        FOR v_source IN EXECUTE pg_catalog.format(
            'SELECT source.* FROM %s AS source '
            'WHERE (coalesce(source.created_at,''-infinity''::timestamp),source.id) > ($1,$2) '
            'ORDER BY coalesce(source.created_at,''-infinity''::timestamp),source.id LIMIT $3',
            v_relation)
            USING v_run.cursor_created_at,v_run.cursor_id,p_limit
        LOOP
            IF v_source.id IS NULL OR v_source.id='00000000-0000-0000-0000-000000000000'::uuid
               OR v_source.entity_id IS NULL OR pg_catalog.octet_length(v_source.entity_id) NOT BETWEEN 1 AND 200 THEN
                RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='AUDIT_V1_SOURCE_ROW_INVALID';
            END IF;
            v_source_document := pg_catalog.to_jsonb(v_source);
            v_source_sha := pg_catalog.sha256(pg_catalog.convert_to(v_source_document::text,'UTF8'));
            v_changes := audit_v1_redacted_changes_v1(
                v_source.old_data,v_source.new_data,v_source.changed_fields);
            v_target_id := audit_v1_derived_id_v1(v_source.id,'entry');
            v_owner_id := audit_v1_derived_id_v1(v_source.id,'owner');
            v_actor_id := CASE WHEN v_source.user_id IS NULL OR
                v_source.user_id='00000000-0000-0000-0000-000000000000'::uuid
                THEN audit_v1_derived_id_v1(
                    '00000000-0000-0000-0000-000000000001'::uuid,'system_actor')
                ELSE v_source.user_id END;
            v_org_id := CASE WHEN v_source.org_id IS NULL OR
                v_source.org_id='00000000-0000-0000-0000-000000000000'::uuid
                THEN NULL ELSE v_source.org_id END;
            v_correlation_id := NULL;
            IF v_source.correlation_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN
                v_correlation_id := v_source.correlation_id::uuid;
                IF v_correlation_id='00000000-0000-0000-0000-000000000000'::uuid THEN
                    v_correlation_id := NULL;
                END IF;
            END IF;
            v_owner_module := audit_v1_owner_module_v1(v_source.entity_type);
            v_operation := audit_v1_operation_v1(v_source.entity_type,v_source.action);
            v_entity_type := CASE WHEN pg_catalog.lower(v_source.entity_type) ~
                '^[a-z][a-z0-9_.-]{0,127}$' THEN pg_catalog.lower(v_source.entity_type)
                ELSE 'legacy_entity' END;
            v_form_slug := CASE WHEN pg_catalog.lower(v_source.form_slug) ~
                '^[a-z][a-z0-9_.-]{0,99}$' THEN pg_catalog.lower(v_source.form_slug)
                ELSE NULL END;
            v_occurred_at := coalesce(v_source.created_at,
                p_observed_at AT TIME ZONE 'UTC') AT TIME ZONE 'UTC';
            INSERT INTO audit_entries (
                id,owner_audit_id,actor_id,organization_id,correlation_id,
                owner_module,operation_id,entity_type,entity_id,form_slug,action,
                authorization_version,reason_code,field_changes,event_sha256,
                occurred_at,recorded_at
            ) VALUES (
                v_target_id,v_owner_id,v_actor_id,v_org_id,v_correlation_id,
                v_owner_module,v_operation,v_entity_type,v_source.entity_id,v_form_slug,
                pg_catalog.lower(v_source.action),1,'v1_adopted',v_changes,v_source_sha,
                v_occurred_at,greatest(p_observed_at,v_occurred_at));
            v_target_sha := pg_catalog.sha256(pg_catalog.convert_to(
                pg_catalog.jsonb_build_object(
                    'id',v_target_id,'owner_audit_id',v_owner_id,'actor_id',v_actor_id,
                    'owner_module',v_owner_module,'operation',v_operation,
                    'entity_type',v_entity_type,'entity_id',v_source.entity_id,
                    'field_changes',v_changes,'event_sha256',pg_catalog.encode(v_source_sha,'hex'))::text,
                'UTF8'));
            INSERT INTO audit_v1_adoption_mappings (
                source_id,run_id,target_entry_id,target_owner_audit_id,source_sha256,
                target_sha256,page_number,adopted_at
            ) VALUES (
                v_source.id,p_run_id,v_target_id,v_owner_id,v_source_sha,
                v_target_sha,v_page,p_observed_at);
            v_source_hex := v_source_hex || pg_catalog.encode(v_source_sha,'hex');
            v_target_hex := v_target_hex || pg_catalog.encode(v_target_sha,'hex');
            v_cursor_at := coalesce(v_source.created_at,'-infinity'::timestamp);
            v_cursor_id := v_source.id;
            v_rows := v_rows + 1;
        END LOOP;
    END IF;

    IF v_rows > 0 THEN
        UPDATE audit_v1_adoption_runs AS run SET
            adopted_count=run.adopted_count+v_rows,next_page=run.next_page+1,
            cursor_created_at=v_cursor_at,cursor_id=v_cursor_id
         WHERE run.run_id=p_run_id RETURNING * INTO v_run;
    END IF;
    IF v_run.adopted_count=v_run.source_count AND v_run.status='applying' THEN
        IF v_relation IS NOT NULL THEN
            EXECUTE pg_catalog.format(
                'SELECT count(*) FROM %s AS source LEFT JOIN audit_v1_adoption_mappings AS mapping '
                'ON mapping.source_id=source.id AND mapping.source_sha256='
                'pg_catalog.sha256(pg_catalog.convert_to(pg_catalog.to_jsonb(source)::text,''UTF8'')) '
                'WHERE mapping.source_id IS NULL',v_relation) INTO v_drift_count;
            EXECUTE pg_catalog.format('SELECT count(*) FROM %s',v_relation) INTO v_source_count;
        ELSE
            v_drift_count := 0;
            v_source_count := 0;
        END IF;
        IF v_drift_count <> 0 OR v_source_count <> v_run.source_count THEN
            RAISE EXCEPTION USING ERRCODE='P6803', MESSAGE='AUDIT_V1_SOURCE_DRIFT';
        END IF;
        UPDATE audit_v1_adoption_runs AS run SET status='completed',completed_at=p_observed_at
         WHERE run.run_id=p_run_id RETURNING * INTO v_run;
    END IF;

    v_response := audit_v1_run_document_v1(v_run);
    INSERT INTO audit_v1_adoption_receipts (
        operation_id,run_id,request_sha256,page_number,row_count,
        source_page_sha256,target_page_sha256,response_document,applied_at
    ) VALUES (
        p_operation_id,p_run_id,v_request_sha,v_page,v_rows,
        pg_catalog.sha256(pg_catalog.convert_to(v_source_hex,'UTF8')),
        pg_catalog.sha256(pg_catalog.convert_to(v_target_hex,'UTF8')),
        v_response,p_observed_at);
    RETURN v_response;
END;
$$;
-- +goose StatementEnd
