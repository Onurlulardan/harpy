-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_assert_v1_adoption_ready_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_observed_source_sha256 text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_rehearsal platform.billing_v1_adoption_rollback_rehearsals%ROWTYPE;
    v_target bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter'
       OR p_observed_source_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 readiness assertion is denied';
    END IF;
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
      AND run.state = 'rollback_rehearsed'
      AND run.money_difference = 0 AND run.period_difference = 0
      AND run.coupon_difference = 0 AND run.usage_difference = 0;
    SELECT rehearsal.* INTO STRICT v_rehearsal
    FROM platform.billing_v1_adoption_rollback_rehearsals AS rehearsal
    WHERE rehearsal.operation_id = p_operation_id AND rehearsal.run_id = p_run_id
      AND rehearsal.zero_difference
      AND rehearsal.observed_source_sha256 =
          pg_catalog.decode(p_observed_source_sha256, 'hex')
      AND rehearsal.reverse_sha256 = rehearsal.observed_source_sha256;
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
    IF v_run.source_sha256 IS DISTINCT FROM
            pg_catalog.decode(p_observed_source_sha256, 'hex')
       OR v_run.target_sha256 IS DISTINCT FROM v_target
       OR v_rehearsal.target_sha256 IS DISTINCT FROM v_target THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing V1 adoption readiness evidence drifted';
    END IF;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_assert_v1_adoption_ready_v1(
    uuid, uuid, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_assert_v1_adoption_ready_v1(
    uuid, uuid, text
) TO harpy_adopter;
-- +goose StatementEnd
