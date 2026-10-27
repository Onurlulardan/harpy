-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.usage_adoption_apply_page_v1(
    p_operation_id uuid, p_run_id uuid, p_lease_owner uuid, p_lease_generation bigint,
    p_page_number bigint, p_cursor_before uuid, p_cursor_after uuid, p_records jsonb,
    p_source_page_sha256 bytea, p_target_page_sha256 bytea,
    p_source_relation_oid oid, p_source_count bigint, p_source_root_sha256 bytea,
    p_observed_at timestamptz
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE
    v_run platform.usage_adoption_runs%ROWTYPE;
    v_receipt platform.usage_adoption_page_receipts%ROWTYPE;
    v_mapping platform.usage_adoption_mappings%ROWTYPE;
    v_record jsonb;
    v_fact_id uuid;
    v_replayed boolean;
    v_source_id uuid;
    v_target_id uuid;
    v_source_sha bytea;
    v_target_sha bytea;
BEGIN
    SELECT * INTO v_run FROM platform.usage_adoption_runs AS runs
    WHERE runs.operation_id = p_operation_id AND runs.run_id = p_run_id FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = 'P6904', MESSAGE = 'usage adoption missing'; END IF;
    IF v_run.source_relation_oid IS DISTINCT FROM p_source_relation_oid
       OR v_run.source_count IS DISTINCT FROM p_source_count
       OR v_run.source_root_sha256 IS DISTINCT FROM p_source_root_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6906', MESSAGE = 'usage adoption source drifted';
    END IF;
    IF v_run.status <> 'applying' OR v_run.lease_owner IS DISTINCT FROM p_lease_owner
       OR v_run.lease_generation IS DISTINCT FROM p_lease_generation
       OR v_run.lease_expires_at <= p_observed_at OR v_run.next_page <> p_page_number
       OR v_run.cursor_id IS DISTINCT FROM p_cursor_before OR p_cursor_after IS NULL
       OR p_records IS NULL OR pg_catalog.jsonb_typeof(p_records) <> 'array'
       OR pg_catalog.jsonb_array_length(p_records) NOT BETWEEN 1 AND 100
       OR pg_catalog.octet_length(p_source_page_sha256) <> 32
       OR pg_catalog.octet_length(p_target_page_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage adoption page conflicts';
    END IF;
    SELECT * INTO v_receipt FROM platform.usage_adoption_page_receipts AS receipts
    WHERE receipts.run_id = p_run_id AND receipts.page_number = p_page_number;
    IF FOUND THEN
        IF v_receipt.cursor_before IS DISTINCT FROM p_cursor_before
           OR v_receipt.cursor_after IS DISTINCT FROM p_cursor_after
           OR v_receipt.row_count <> pg_catalog.jsonb_array_length(p_records)
           OR v_receipt.source_page_sha256 IS DISTINCT FROM p_source_page_sha256
           OR v_receipt.target_page_sha256 IS DISTINCT FROM p_target_page_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage adoption receipt conflicts';
        END IF;
        RETURN pg_catalog.to_jsonb(v_receipt);
    END IF;
    FOR v_record IN SELECT value FROM pg_catalog.jsonb_array_elements(p_records) AS rows(value) LOOP
        IF pg_catalog.jsonb_typeof(v_record) <> 'object'
           OR (v_record - ARRAY['source_id','target_id','environment_id','tenant_id','actor_id',
              'source_identity','source_sha256','fact_digest','method','route_pattern','status_code',
              'duration_micros','occurred_at']::text[]) <> '{}'::jsonb THEN
            RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage adoption record is invalid';
        END IF;
        v_source_id := (v_record->>'source_id')::uuid;
        v_target_id := (v_record->>'target_id')::uuid;
        v_source_sha := pg_catalog.decode(v_record->>'source_sha256', 'hex');
        v_target_sha := pg_catalog.decode(v_record->>'fact_digest', 'hex');
        SELECT recorded.fact_id, recorded.replayed INTO v_fact_id, v_replayed
        FROM platform.usage_record_fact_v1(v_target_id,
            (v_record->>'environment_id')::uuid, (v_record->>'tenant_id')::uuid,
            (v_record->>'actor_id')::uuid,
            pg_catalog.decode(v_record->>'source_identity', 'hex'), v_target_sha,
            v_record->>'method', v_record->>'route_pattern',
            (v_record->>'status_code')::integer, (v_record->>'duration_micros')::bigint,
            (v_record->>'occurred_at')::timestamptz) AS recorded;
        IF v_fact_id IS DISTINCT FROM v_target_id OR pg_catalog.octet_length(v_source_sha) <> 32 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage adoption mapping conflicts';
        END IF;
        INSERT INTO platform.usage_adoption_mappings AS target (
            run_id, source_id, target_id, source_sha256, target_sha256, page_number, applied_at
        ) VALUES (p_run_id, v_source_id, v_target_id, v_source_sha,
            v_target_sha, p_page_number, p_observed_at)
        ON CONFLICT (run_id, source_id) DO NOTHING;
        SELECT * INTO v_mapping FROM platform.usage_adoption_mappings AS mappings
        WHERE mappings.run_id = p_run_id AND mappings.source_id = v_source_id;
        IF v_mapping.target_id IS DISTINCT FROM v_target_id
           OR v_mapping.source_sha256 IS DISTINCT FROM v_source_sha
           OR v_mapping.target_sha256 IS DISTINCT FROM v_target_sha
           OR v_mapping.page_number IS DISTINCT FROM p_page_number THEN
            RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage adoption mapping drifted';
        END IF;
    END LOOP;
    INSERT INTO platform.usage_adoption_page_receipts (
        run_id, page_number, cursor_before, cursor_after, row_count,
        source_page_sha256, target_page_sha256, state, applied_at
    ) VALUES (p_run_id, p_page_number, p_cursor_before, p_cursor_after,
        pg_catalog.jsonb_array_length(p_records), p_source_page_sha256,
        p_target_page_sha256, 'applied', p_observed_at)
    RETURNING * INTO v_receipt;
    RETURN pg_catalog.to_jsonb(v_receipt);
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_adoption_apply_page_v1(
uuid,uuid,uuid,bigint,bigint,uuid,uuid,jsonb,bytea,bytea,oid,bigint,bytea,timestamptz)
FROM PUBLIC;
-- +goose StatementEnd
