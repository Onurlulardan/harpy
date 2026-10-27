-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.support_adoption_target_drift_v1(p_run_id uuid)
RETURNS boolean
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
    SELECT EXISTS (
        SELECT 1
        FROM platform.support_adoption_mappings AS mapping
        WHERE mapping.run_id=p_run_id
          AND mapping.target_sha256 IS DISTINCT FROM pg_catalog.sha256(
              pg_catalog.convert_to((
                  CASE mapping.source_kind
                  WHEN 'contact_global' THEN
                      platform.support_contact_adoption_document_v1(mapping.target_id)
                  WHEN 'bug_tenant' THEN
                      platform.support_bug_adoption_document_v1(mapping.target_id)
                  END
              )::text,'UTF8')
          )
    )
$$;

CREATE FUNCTION platform.support_reconcile_adoption_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_lease_owner uuid,
    p_lease_generation bigint,
    p_observed_at timestamp(6) with time zone
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
DECLARE
    v_run platform.support_adoption_runs%ROWTYPE;
    v_mapping_count bigint;
    v_page_count bigint;
BEGIN
    PERFORM platform.support_require_adopter_v1();
    IF p_observed_at IS NULL
       OR pg_catalog.date_trunc('microseconds',p_observed_at) <> p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='invalid support adoption reconcile';
    END IF;
    SELECT run.* INTO v_run FROM platform.support_adoption_runs AS run
    WHERE run.operation_id=p_operation_id AND run.run_id=p_run_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6514', MESSAGE='support adoption run not found';
    END IF;
    IF v_run.lease_owner <> p_lease_owner
       OR v_run.lease_generation <> p_lease_generation
       OR v_run.lease_expires_at <= p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='P6513', MESSAGE='support adoption lease mismatch';
    END IF;
    IF v_run.status NOT IN ('applying','reconciling') THEN
        RAISE EXCEPTION USING ERRCODE='P6511', MESSAGE='support adoption reconcile state mismatch';
    END IF;
    SELECT pg_catalog.count(*) INTO v_mapping_count
    FROM platform.support_adoption_mappings AS mapping WHERE mapping.run_id=p_run_id;
    SELECT COALESCE(pg_catalog.sum(page.row_count),0::bigint) INTO v_page_count
    FROM platform.support_adoption_page_receipts AS page
    WHERE page.run_id=p_run_id AND page.state='advanced';
    IF v_mapping_count <> v_run.copied_count OR v_page_count <> v_run.copied_count
       OR v_run.copied_count <> v_run.expected_count
       OR v_run.source_sha256 <> v_run.expected_source_sha256
       OR EXISTS (SELECT 1 FROM platform.support_adoption_page_receipts AS page
                  WHERE page.run_id=p_run_id AND page.state<>'advanced')
       OR platform.support_adoption_target_drift_v1(p_run_id) THEN
        RAISE EXCEPTION USING ERRCODE='P6512', MESSAGE='support adoption reconciliation drift';
    END IF;
    IF v_run.status='reconciling' THEN
        RETURN platform.support_adoption_run_document_v1(v_run);
    END IF;
    UPDATE platform.support_adoption_runs AS run SET
        status='reconciling',reconciled_at=p_observed_at,
        revision=run.revision+1,lease_expires_at=p_observed_at+interval '5 minutes',
        updated_at=p_observed_at
    WHERE run.run_id=p_run_id RETURNING * INTO v_run;
    RETURN platform.support_adoption_run_document_v1(v_run);
END;
$$;

CREATE FUNCTION platform.support_complete_adoption_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_lease_owner uuid,
    p_lease_generation bigint,
    p_observed_at timestamp(6) with time zone
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
DECLARE
    v_run platform.support_adoption_runs%ROWTYPE;
BEGIN
    PERFORM platform.support_require_adopter_v1();
    IF p_observed_at IS NULL
       OR pg_catalog.date_trunc('microseconds',p_observed_at) <> p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='invalid support adoption completion';
    END IF;
    SELECT run.* INTO v_run FROM platform.support_adoption_runs AS run
    WHERE run.operation_id=p_operation_id AND run.run_id=p_run_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6514', MESSAGE='support adoption run not found';
    END IF;
    IF v_run.lease_owner <> p_lease_owner OR v_run.lease_generation <> p_lease_generation THEN
        RAISE EXCEPTION USING ERRCODE='P6513', MESSAGE='support adoption lease mismatch';
    END IF;
    IF v_run.status='completed' THEN
        RETURN platform.support_adoption_run_document_v1(v_run);
    END IF;
    IF v_run.status <> 'reconciling' OR v_run.reconciled_at IS NULL
       OR v_run.lease_expires_at <= p_observed_at
       OR v_run.copied_count <> v_run.expected_count
       OR v_run.source_sha256 <> v_run.expected_source_sha256
       OR platform.support_adoption_target_drift_v1(p_run_id) THEN
        RAISE EXCEPTION USING ERRCODE='P6512', MESSAGE='support adoption completion drift';
    END IF;
    UPDATE platform.support_adoption_runs AS run SET
        status='completed',completed_at=p_observed_at,lease_expires_at=p_observed_at,
        revision=run.revision+1,updated_at=p_observed_at
    WHERE run.run_id=p_run_id RETURNING * INTO v_run;
    RETURN platform.support_adoption_run_document_v1(v_run);
END;
$$;

REVOKE ALL ON FUNCTION platform.support_adoption_target_drift_v1(uuid) FROM PUBLIC, harpy_runtime, harpy_adopter;
REVOKE ALL ON FUNCTION platform.support_reconcile_adoption_v1(uuid,uuid,uuid,bigint,timestamp with time zone) FROM PUBLIC, harpy_runtime;
REVOKE ALL ON FUNCTION platform.support_complete_adoption_v1(uuid,uuid,uuid,bigint,timestamp with time zone) FROM PUBLIC, harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.support_reconcile_adoption_v1(uuid,uuid,uuid,bigint,timestamp with time zone) TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.support_complete_adoption_v1(uuid,uuid,uuid,bigint,timestamp with time zone) TO harpy_adopter;
-- +goose StatementEnd
