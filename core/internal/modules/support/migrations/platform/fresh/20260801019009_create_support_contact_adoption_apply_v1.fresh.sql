-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.support_apply_contact_adoption_page_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_lease_owner uuid,
    p_lease_generation bigint,
    p_page_number bigint,
    p_cursor_before uuid,
    p_rows jsonb,
    p_source_page_sha256 bytea,
    p_observed_at timestamp(6) with time zone
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
DECLARE
    v_run platform.support_adoption_runs%ROWTYPE;
    v_page platform.support_adoption_page_receipts%ROWTYPE;
    v_item jsonb;
    v_target jsonb;
    v_source_id uuid;
    v_target_id uuid;
    v_previous uuid := p_cursor_before;
    v_source_hex text := '';
    v_target_hex text := '';
    v_source_sha bytea;
    v_target_sha bytea;
    v_request_sha bytea;
    v_count integer;
BEGIN
    PERFORM platform.support_require_adopter_v1();
    IF p_page_number < 1 OR pg_catalog.jsonb_typeof(p_rows) <> 'array'
       OR pg_catalog.octet_length(p_source_page_sha256) <> 32
       OR p_observed_at IS NULL
       OR pg_catalog.date_trunc('microseconds',p_observed_at) <> p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='invalid support contact adoption page';
    END IF;
    v_count := pg_catalog.jsonb_array_length(p_rows);
    IF v_count < 1 OR v_count > 100 THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='support adoption page is outside budget';
    END IF;
    v_request_sha := pg_catalog.sha256(pg_catalog.convert_to(p_rows::text,'UTF8'));

    SELECT run.* INTO v_run FROM platform.support_adoption_runs AS run
    WHERE run.operation_id=p_operation_id AND run.run_id=p_run_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6514', MESSAGE='support adoption run not found';
    END IF;
    IF v_run.source_kind <> 'contact_global' THEN
        RAISE EXCEPTION USING ERRCODE='P6511', MESSAGE='support adoption source mismatch';
    END IF;
    IF v_run.lease_owner <> p_lease_owner
       OR v_run.lease_generation <> p_lease_generation
       OR v_run.lease_expires_at <= p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='P6513', MESSAGE='support adoption lease mismatch';
    END IF;

    SELECT page.* INTO v_page FROM platform.support_adoption_page_receipts AS page
    WHERE page.run_id=p_run_id AND page.page_number=p_page_number;
    IF FOUND THEN
        IF v_page.cursor_before IS DISTINCT FROM p_cursor_before
           OR v_page.row_count <> v_count OR v_page.request_sha256 <> v_request_sha
           OR v_page.source_page_sha256 <> p_source_page_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='P6512', MESSAGE='support adoption page drift';
        END IF;
        RETURN platform.support_adoption_page_document_v1(v_page);
    END IF;

    IF v_run.status <> 'applying'
       OR v_run.next_page <> p_page_number
       OR v_run.cursor_id IS DISTINCT FROM p_cursor_before THEN
        RAISE EXCEPTION USING ERRCODE='P6511', MESSAGE='support adoption page state mismatch';
    END IF;

    FOR v_item IN SELECT value FROM pg_catalog.jsonb_array_elements(p_rows)
    LOOP
        IF pg_catalog.jsonb_typeof(v_item) <> 'object'
           OR NOT (v_item ?& ARRAY['source_id','source_sha256','target'])
           OR (v_item - ARRAY['source_id','source_sha256','target']) <> '{}'::jsonb
           OR (v_item->>'source_sha256') !~ '^[0-9a-f]{64}$'
           OR pg_catalog.jsonb_typeof(v_item->'target') <> 'object' THEN
            RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='invalid support contact adoption record';
        END IF;
        v_source_id := (v_item->>'source_id')::uuid;
        IF v_source_id='00000000-0000-0000-0000-000000000000'
           OR (v_previous IS NOT NULL AND v_source_id <= v_previous) THEN
            RAISE EXCEPTION USING ERRCODE='P6512', MESSAGE='support source order drift';
        END IF;
        v_target := v_item->'target';
        IF NOT (v_target ?& ARRAY['name','email','company','phone','employee_count',
            'message','contact_type','status','source','notes','replied_at','replied_by',
            'created_at','updated_at'])
           OR (v_target - ARRAY['name','email','company','phone','employee_count',
            'message','contact_type','status','source','notes','replied_at','replied_by',
            'created_at','updated_at']) <> '{}'::jsonb THEN
            RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='invalid support contact target';
        END IF;

        v_target_id := pg_catalog.uuidv7();
        INSERT INTO platform.support_contact_messages (
            id,name,email,company,phone,employee_count,message,contact_type,status,
            source,notes,replied_at,replied_by,revision,created_at,updated_at
        ) VALUES (
            v_target_id,v_target->>'name',v_target->>'email',v_target->>'company',
            v_target->>'phone',v_target->>'employee_count',v_target->>'message',
            v_target->>'contact_type',v_target->>'status',v_target->>'source',
            v_target->>'notes',(v_target->>'replied_at')::timestamp with time zone,
            (v_target->>'replied_by')::uuid,1,
            (v_target->>'created_at')::timestamp with time zone,
            (v_target->>'updated_at')::timestamp with time zone
        );
        v_source_sha := pg_catalog.decode(v_item->>'source_sha256','hex');
        v_target_sha := pg_catalog.sha256(pg_catalog.convert_to(
            platform.support_contact_adoption_document_v1(v_target_id)::text,'UTF8'));
        INSERT INTO platform.support_adoption_mappings (
            run_id,source_kind,source_tenant_id,source_id,target_id,source_sha256,
            target_sha256,page_number,applied_at
        ) VALUES (
            p_run_id,'contact_global',NULL,v_source_id,v_target_id,v_source_sha,
            v_target_sha,p_page_number,p_observed_at
        );
        v_source_hex := v_source_hex || pg_catalog.encode(v_source_sha,'hex');
        v_target_hex := v_target_hex || pg_catalog.encode(v_target_sha,'hex');
        v_previous := v_source_id;
    END LOOP;

    v_source_sha := pg_catalog.sha256(pg_catalog.convert_to(v_source_hex,'UTF8'));
    v_target_sha := pg_catalog.sha256(pg_catalog.convert_to(v_target_hex,'UTF8'));
    IF v_source_sha <> p_source_page_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='P6512', MESSAGE='support source page checksum drift';
    END IF;
    INSERT INTO platform.support_adoption_page_receipts (
        run_id,page_number,cursor_before,cursor_after,row_count,request_sha256,
        source_page_sha256,target_page_sha256,state,applied_at
    ) VALUES (
        p_run_id,p_page_number,p_cursor_before,v_previous,v_count,v_request_sha,
        v_source_sha,v_target_sha,'applied',p_observed_at
    ) RETURNING * INTO v_page;
    RETURN platform.support_adoption_page_document_v1(v_page);
END;
$$;

REVOKE ALL ON FUNCTION platform.support_apply_contact_adoption_page_v1(uuid,uuid,uuid,bigint,bigint,uuid,jsonb,bytea,timestamp with time zone) FROM PUBLIC, forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.support_apply_contact_adoption_page_v1(uuid,uuid,uuid,bigint,bigint,uuid,jsonb,bytea,timestamp with time zone) TO forgecrud_adopter;
-- +goose StatementEnd
