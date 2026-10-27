-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.support_advance_adoption_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_lease_owner uuid,
    p_lease_generation bigint,
    p_page_number bigint,
    p_cursor_before uuid,
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
    v_source_root bytea;
    v_target_root bytea;
BEGIN
    PERFORM platform.support_require_adopter_v1();
    IF p_page_number < 1 OR pg_catalog.octet_length(p_source_page_sha256) <> 32
       OR p_observed_at IS NULL
       OR pg_catalog.date_trunc('microseconds',p_observed_at) <> p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='invalid support adoption advance';
    END IF;
    SELECT run.* INTO v_run FROM platform.support_adoption_runs AS run
    WHERE run.operation_id=p_operation_id AND run.run_id=p_run_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6514', MESSAGE='support adoption run not found';
    END IF;
    SELECT page.* INTO v_page FROM platform.support_adoption_page_receipts AS page
    WHERE page.run_id=p_run_id AND page.page_number=p_page_number FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6514', MESSAGE='support adoption page not found';
    END IF;
    IF v_page.cursor_before IS DISTINCT FROM p_cursor_before
       OR v_page.source_page_sha256 <> p_source_page_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='P6512', MESSAGE='support adoption advance drift';
    END IF;
    IF v_page.state='advanced' THEN
        IF v_run.next_page <> p_page_number+1 OR v_run.cursor_id <> v_page.cursor_after THEN
            RAISE EXCEPTION USING ERRCODE='P6512', MESSAGE='support adoption checkpoint drift';
        END IF;
        RETURN platform.support_adoption_run_document_v1(v_run);
    END IF;
    IF v_run.status <> 'applying' OR v_run.next_page <> p_page_number
       OR v_run.cursor_id IS DISTINCT FROM p_cursor_before
       OR v_run.copied_count+v_page.row_count > v_run.expected_count THEN
        RAISE EXCEPTION USING ERRCODE='P6511', MESSAGE='support adoption advance state mismatch';
    END IF;
    IF v_run.lease_owner <> p_lease_owner
       OR v_run.lease_generation <> p_lease_generation
       OR v_run.lease_expires_at <= p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='P6513', MESSAGE='support adoption lease mismatch';
    END IF;

    v_source_root := pg_catalog.sha256(pg_catalog.convert_to(
        pg_catalog.encode(v_run.source_sha256,'hex') ||
        pg_catalog.encode(v_page.source_page_sha256,'hex'),'UTF8'));
    v_target_root := pg_catalog.sha256(pg_catalog.convert_to(
        pg_catalog.encode(v_run.target_sha256,'hex') ||
        pg_catalog.encode(v_page.target_page_sha256,'hex'),'UTF8'));
    UPDATE platform.support_adoption_page_receipts AS page
    SET state='advanced',advanced_at=p_observed_at
    WHERE page.run_id=p_run_id AND page.page_number=p_page_number
    RETURNING * INTO v_page;
    UPDATE platform.support_adoption_runs AS run SET
        cursor_id=v_page.cursor_after,copied_count=run.copied_count+v_page.row_count,
        source_sha256=v_source_root,target_sha256=v_target_root,
        next_page=run.next_page+1,revision=run.revision+1,
        lease_expires_at=p_observed_at+interval '5 minutes',updated_at=p_observed_at
    WHERE run.run_id=p_run_id RETURNING * INTO v_run;
    RETURN platform.support_adoption_run_document_v1(v_run);
END;
$$;

REVOKE ALL ON FUNCTION platform.support_advance_adoption_v1(uuid,uuid,uuid,bigint,bigint,uuid,bytea,timestamp with time zone) FROM PUBLIC, forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.support_advance_adoption_v1(uuid,uuid,uuid,bigint,bigint,uuid,bytea,timestamp with time zone) TO forgecrud_adopter;
-- +goose StatementEnd

