-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.usage_adoption_advance_v1(
    p_operation_id uuid, p_run_id uuid, p_lease_owner uuid, p_lease_generation bigint,
    p_page_number bigint, p_cursor_before uuid, p_cursor_after uuid,
    p_source_page_sha256 bytea, p_target_page_sha256 bytea,
    p_source_relation_oid oid, p_source_count bigint, p_source_root_sha256 bytea,
    p_observed_at timestamptz
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE v_run platform.usage_adoption_runs%ROWTYPE; v_receipt platform.usage_adoption_page_receipts%ROWTYPE;
BEGIN
    SELECT * INTO v_run FROM platform.usage_adoption_runs AS runs
    WHERE runs.operation_id = p_operation_id AND runs.run_id = p_run_id FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = 'P6904', MESSAGE = 'usage adoption missing'; END IF;
    IF v_run.source_relation_oid IS DISTINCT FROM p_source_relation_oid
       OR v_run.source_count IS DISTINCT FROM p_source_count
       OR v_run.source_root_sha256 IS DISTINCT FROM p_source_root_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6906', MESSAGE = 'usage adoption source drifted';
    END IF;
    SELECT * INTO v_receipt FROM platform.usage_adoption_page_receipts AS receipts
    WHERE receipts.run_id = p_run_id AND receipts.page_number = p_page_number FOR UPDATE;
    IF NOT FOUND OR v_receipt.cursor_before IS DISTINCT FROM p_cursor_before
       OR v_receipt.cursor_after IS DISTINCT FROM p_cursor_after
       OR v_receipt.source_page_sha256 IS DISTINCT FROM p_source_page_sha256
       OR v_receipt.target_page_sha256 IS DISTINCT FROM p_target_page_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage adoption advance conflicts';
    END IF;
    IF v_receipt.state = 'advanced' AND v_run.next_page = p_page_number + 1
       AND v_run.cursor_id = p_cursor_after THEN
        RETURN platform.usage_adoption_run_document_v1(p_run_id);
    END IF;
    IF v_receipt.state <> 'applied' OR v_run.status <> 'applying'
       OR v_run.lease_owner IS DISTINCT FROM p_lease_owner
       OR v_run.lease_generation IS DISTINCT FROM p_lease_generation
       OR v_run.lease_expires_at <= p_observed_at OR v_run.next_page <> p_page_number
       OR v_run.cursor_id IS DISTINCT FROM p_cursor_before THEN
        RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage adoption lease conflicts';
    END IF;
    UPDATE platform.usage_adoption_page_receipts SET state = 'advanced', advanced_at = p_observed_at
    WHERE run_id = p_run_id AND page_number = p_page_number;
    UPDATE platform.usage_adoption_runs SET next_page = next_page + 1,
        cursor_id = p_cursor_after, copied_count = copied_count + v_receipt.row_count,
        revision = revision + 1, updated_at = p_observed_at,
        lease_expires_at = p_observed_at + interval '30 seconds' WHERE run_id = p_run_id;
    RETURN platform.usage_adoption_run_document_v1(p_run_id);
END;
$$;

CREATE FUNCTION platform.usage_adoption_reconcile_v1(
    p_operation_id uuid, p_run_id uuid, p_lease_owner uuid, p_lease_generation bigint,
    p_source_relation_oid oid, p_source_count bigint, p_source_root_sha256 bytea,
    p_observed_at timestamptz
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE v_run platform.usage_adoption_runs%ROWTYPE; v_mapping_count bigint;
BEGIN
    SELECT * INTO v_run FROM platform.usage_adoption_runs AS runs
    WHERE runs.operation_id = p_operation_id AND runs.run_id = p_run_id FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = 'P6904', MESSAGE = 'usage adoption missing'; END IF;
    IF v_run.source_relation_oid IS DISTINCT FROM p_source_relation_oid
       OR v_run.source_count IS DISTINCT FROM p_source_count
       OR v_run.source_root_sha256 IS DISTINCT FROM p_source_root_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6906', MESSAGE = 'usage adoption source drifted';
    END IF;
    SELECT pg_catalog.count(*) INTO v_mapping_count FROM platform.usage_adoption_mappings AS mappings
    WHERE mappings.run_id = p_run_id;
    IF v_run.status <> 'applying' OR v_run.lease_owner IS DISTINCT FROM p_lease_owner
       OR v_run.lease_generation IS DISTINCT FROM p_lease_generation
       OR v_run.lease_expires_at <= p_observed_at OR v_run.copied_count <> p_source_count
       OR v_mapping_count <> p_source_count THEN
        RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage adoption reconcile conflicts';
    END IF;
    UPDATE platform.usage_adoption_runs SET status = 'reconciling', revision = revision + 1,
        updated_at = p_observed_at, lease_expires_at = p_observed_at + interval '30 seconds'
    WHERE run_id = p_run_id;
    RETURN platform.usage_adoption_run_document_v1(p_run_id);
END;
$$;

CREATE FUNCTION platform.usage_adoption_complete_v1(
    p_operation_id uuid, p_run_id uuid, p_lease_owner uuid, p_lease_generation bigint,
    p_source_relation_oid oid, p_source_count bigint, p_source_root_sha256 bytea,
    p_observed_at timestamptz
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE v_run platform.usage_adoption_runs%ROWTYPE;
BEGIN
    SELECT * INTO v_run FROM platform.usage_adoption_runs AS runs
    WHERE runs.operation_id = p_operation_id AND runs.run_id = p_run_id FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = 'P6904', MESSAGE = 'usage adoption missing'; END IF;
    IF v_run.status = 'completed' THEN RETURN platform.usage_adoption_run_document_v1(p_run_id); END IF;
    IF v_run.source_relation_oid IS DISTINCT FROM p_source_relation_oid
       OR v_run.source_count IS DISTINCT FROM p_source_count
       OR v_run.source_root_sha256 IS DISTINCT FROM p_source_root_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6906', MESSAGE = 'usage adoption source drifted';
    END IF;
    IF v_run.status <> 'reconciling' OR v_run.lease_owner IS DISTINCT FROM p_lease_owner
       OR v_run.lease_generation IS DISTINCT FROM p_lease_generation
       OR v_run.lease_expires_at <= p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage adoption completion conflicts';
    END IF;
    UPDATE platform.usage_adoption_runs SET status = 'completed', completed_at = p_observed_at,
        lease_owner = NULL, lease_expires_at = NULL, revision = revision + 1,
        updated_at = p_observed_at WHERE run_id = p_run_id;
    RETURN platform.usage_adoption_run_document_v1(p_run_id);
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_adoption_advance_v1(
uuid,uuid,uuid,bigint,bigint,uuid,uuid,bytea,bytea,oid,bigint,bytea,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.usage_adoption_reconcile_v1(
uuid,uuid,uuid,bigint,oid,bigint,bytea,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.usage_adoption_complete_v1(
uuid,uuid,uuid,bigint,oid,bigint,bytea,timestamptz) FROM PUBLIC;
-- +goose StatementEnd
