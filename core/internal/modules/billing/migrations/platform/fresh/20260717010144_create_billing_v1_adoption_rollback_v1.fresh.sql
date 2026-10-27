-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_rehearse_v1_adoption_rollback_v1(
    p_operation_id uuid, p_run_id uuid, p_receipt_id uuid,
    p_observed_source_sha256 text, p_rehearsed_at timestamptz
)
RETURNS TABLE (
    receipt_id uuid, source_sha256 text,
    reverse_sha256 text, rehearsed_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_target bytea;
    v_existing platform.billing_v1_adoption_rollback_rehearsals%ROWTYPE;
BEGIN
    IF p_observed_source_sha256 !~ '^[0-9a-f]{64}$'
       OR p_rehearsed_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing V1 rollback observation is invalid';
    END IF;
    SELECT rehearsal.* INTO v_existing
    FROM platform.billing_v1_adoption_rollback_rehearsals AS rehearsal
    WHERE rehearsal.operation_id = p_operation_id AND rehearsal.run_id = p_run_id;
    IF FOUND THEN
        IF v_existing.receipt_id IS DISTINCT FROM p_receipt_id
           OR v_existing.observed_source_sha256 IS DISTINCT FROM
                pg_catalog.decode(p_observed_source_sha256, 'hex') THEN
            RAISE EXCEPTION USING ERRCODE = '22000',
                MESSAGE = 'billing V1 rollback receipt identity was reused';
        END IF;
        RETURN QUERY SELECT v_existing.receipt_id,
            pg_catalog.encode(v_existing.observed_source_sha256, 'hex'),
            pg_catalog.encode(v_existing.reverse_sha256, 'hex'),
            v_existing.rehearsed_at;
        RETURN;
    END IF;
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
      AND run.state = 'verified'
      AND run.money_difference = 0 AND run.period_difference = 0
      AND run.coupon_difference = 0 AND run.usage_difference = 0
    FOR UPDATE;
    IF v_run.source_sha256 IS DISTINCT FROM
       pg_catalog.decode(p_observed_source_sha256, 'hex') THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing V1 source changed before rollback rehearsal';
    END IF;
    SELECT pg_catalog.sha256(pg_catalog.convert_to(COALESCE(pg_catalog.string_agg(
        record_kind || ':' || source_key || ':' ||
        pg_catalog.encode(current_sha, 'hex'), '|'
        ORDER BY record_kind, source_key
    ) FILTER (WHERE current_sha IS NOT NULL), ''), 'UTF8'))
    INTO v_target
    FROM (
        SELECT mapping.record_kind, mapping.source_key,
            platform.billing_v1_adoption_fingerprint_v1(
                mapping.record_kind, mapping.target_id
            ) AS current_sha
        FROM platform.billing_v1_adoption_mappings AS mapping
        WHERE mapping.operation_id = p_operation_id AND mapping.run_id = p_run_id
    ) AS current_rows;
    IF v_target IS DISTINCT FROM v_run.target_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing V2 target changed before rollback rehearsal';
    END IF;
    INSERT INTO platform.billing_v1_adoption_rollback_rehearsals (
        receipt_id, operation_id, run_id, scope_kind, scope_id, authority_id,
        observed_source_sha256, reverse_sha256, target_sha256,
        zero_difference, rehearsed_at
    ) VALUES (
        p_receipt_id, p_operation_id, p_run_id, v_run.scope_kind,
        v_run.scope_id, v_run.authority_id,
        pg_catalog.decode(p_observed_source_sha256, 'hex'),
        pg_catalog.decode(p_observed_source_sha256, 'hex'),
        v_target, true, p_rehearsed_at
    );
    UPDATE platform.billing_v1_adoption_runs SET
        state = 'rollback_rehearsed', rehearsed_at = p_rehearsed_at,
        updated_at = p_rehearsed_at
    WHERE operation_id = p_operation_id AND run_id = p_run_id;
    RETURN QUERY SELECT p_receipt_id, p_observed_source_sha256,
        p_observed_source_sha256, p_rehearsed_at;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_rehearse_v1_adoption_rollback_v1(
    uuid, uuid, uuid, text, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_rehearse_v1_adoption_rollback_v1(
    uuid, uuid, uuid, text, timestamptz
) TO forgecrud_adopter;
-- +goose StatementEnd
