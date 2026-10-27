-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_record_billing_forms_rehearsal_v1(
    p_receipt_id uuid,
    p_tenant_id uuid,
    p_generation bigint,
    p_billing_operation_id uuid,
    p_billing_run_id uuid,
    p_billing_rollback_receipt_id uuid,
    p_billing_source_count bigint,
    p_billing_target_count bigint,
    p_billing_acceptance_sha256 text,
    p_billing_rollback_sha256 text,
    p_forms_operation_id uuid,
    p_forms_run_id uuid,
    p_forms_rollback_receipt_id uuid,
    p_forms_source_count bigint,
    p_forms_target_count bigint,
    p_forms_acceptance_sha256 text,
    p_forms_rollback_sha256 text,
    p_group_sha256 text,
    p_recorded_at timestamptz
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_group platform.cutover_groups%ROWTYPE;
    v_members jsonb;
    v_locked_count integer;
    v_expected_sha256 text;
    v_persisted_sha256 text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       p_generation < 1 OR NOT pg_catalog.isfinite(p_recorded_at) OR
       p_billing_source_count < 0 OR
       p_billing_source_count IS DISTINCT FROM p_billing_target_count OR
       p_forms_source_count < 0 OR
       p_forms_source_count IS DISTINCT FROM p_forms_target_count OR
       p_billing_acceptance_sha256 !~ '^[0-9a-f]{64}$' OR
       p_billing_rollback_sha256 !~ '^[0-9a-f]{64}$' OR
       p_forms_acceptance_sha256 !~ '^[0-9a-f]{64}$' OR
       p_forms_rollback_sha256 !~ '^[0-9a-f]{64}$' OR
       p_group_sha256 !~ '^[0-9a-f]{64}$' OR
       (SELECT pg_catalog.count(DISTINCT identity)
        FROM pg_catalog.unnest(ARRAY[
            p_receipt_id, p_billing_operation_id, p_billing_run_id,
            p_billing_rollback_receipt_id, p_forms_operation_id,
            p_forms_run_id, p_forms_rollback_receipt_id
        ]) AS identity) IS DISTINCT FROM 7 OR
       p_tenant_id = '00000000-0000-0000-0000-000000000000'::uuid THEN
        RAISE EXCEPTION USING ERRCODE='P57B0',
            MESSAGE='billing Forms rehearsal input denied';
    END IF;

    SELECT cutover_group.* INTO STRICT v_group
    FROM platform.cutover_groups AS cutover_group
    WHERE cutover_group.scope_kind='tenant'
      AND cutover_group.scope_id=p_tenant_id
      AND cutover_group.group_id='billing_forms'
    FOR SHARE;
    IF v_group.state IS DISTINCT FROM 'shadow' OR
       v_group.writer IS DISTINCT FROM 'v1' OR
       v_group.generation IS DISTINCT FROM p_generation THEN
        RAISE EXCEPTION USING ERRCODE='P57B1',
            MESSAGE='billing Forms rehearsal generation differs';
    END IF;

    PERFORM 1
    FROM platform.cutover_group_members AS member
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind=member.scope_kind
     AND domain_row.scope_id=member.scope_id
     AND domain_row.domain=member.domain
    WHERE member.scope_kind=v_group.scope_kind
      AND member.scope_id=v_group.scope_id
      AND member.group_id=v_group.group_id
    ORDER BY member.ordinal
    FOR SHARE OF member,domain_row;
    GET DIAGNOSTICS v_locked_count = ROW_COUNT;
    SELECT pg_catalog.jsonb_agg(member.domain ORDER BY member.ordinal)
    INTO v_members
    FROM platform.cutover_group_members AS member
    WHERE member.scope_kind=v_group.scope_kind
      AND member.scope_id=v_group.scope_id
      AND member.group_id=v_group.group_id;
    IF v_locked_count IS DISTINCT FROM 2 OR
       v_members IS DISTINCT FROM '["billing","forms"]'::jsonb OR EXISTS (
        SELECT 1
        FROM platform.cutover_group_members AS member
        JOIN platform.domain_cutovers AS domain_row
          ON domain_row.scope_kind=member.scope_kind
         AND domain_row.scope_id=member.scope_id
         AND domain_row.domain=member.domain
        WHERE member.scope_kind=v_group.scope_kind
          AND member.scope_id=v_group.scope_id
          AND member.group_id=v_group.group_id
          AND (domain_row.state IS DISTINCT FROM v_group.state OR
               domain_row.writer IS DISTINCT FROM v_group.writer OR
               domain_row.generation IS DISTINCT FROM v_group.generation)
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P57B2',
            MESSAGE='billing Forms rehearsal members differ';
    END IF;

    v_expected_sha256 := platform.cutover_billing_forms_rehearsal_sha256_v1(
        p_tenant_id,p_generation,
        p_billing_operation_id,p_billing_run_id,p_billing_rollback_receipt_id,
        p_billing_source_count,p_billing_target_count,
        p_billing_acceptance_sha256,p_billing_rollback_sha256,
        p_forms_operation_id,p_forms_run_id,p_forms_rollback_receipt_id,
        p_forms_source_count,p_forms_target_count,
        p_forms_acceptance_sha256,p_forms_rollback_sha256
    );
    IF v_expected_sha256 IS DISTINCT FROM p_group_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='P57B3',
            MESSAGE='billing Forms rehearsal digest differs';
    END IF;

    INSERT INTO platform.cutover_billing_forms_rehearsals (
        receipt_id,scope_kind,tenant_id,group_id,group_generation,
        billing_operation_id,billing_run_id,billing_rollback_receipt_id,
        billing_source_count,billing_target_count,billing_difference_count,
        billing_acceptance_sha256,billing_rollback_sha256,
        forms_operation_id,forms_run_id,forms_rollback_receipt_id,
        forms_source_count,forms_target_count,forms_difference_count,
        forms_acceptance_sha256,forms_rollback_sha256,
        quiesce_drained,usage_flushed,production_dual_write,
        group_sha256,recorded_at
    ) VALUES (
        p_receipt_id,'tenant',p_tenant_id,'billing_forms',p_generation,
        p_billing_operation_id,p_billing_run_id,p_billing_rollback_receipt_id,
        p_billing_source_count,p_billing_target_count,0,
        pg_catalog.decode(p_billing_acceptance_sha256,'hex'),
        pg_catalog.decode(p_billing_rollback_sha256,'hex'),
        p_forms_operation_id,p_forms_run_id,p_forms_rollback_receipt_id,
        p_forms_source_count,p_forms_target_count,0,
        pg_catalog.decode(p_forms_acceptance_sha256,'hex'),
        pg_catalog.decode(p_forms_rollback_sha256,'hex'),
        true,true,false,pg_catalog.decode(p_group_sha256,'hex'),p_recorded_at
    ) ON CONFLICT DO NOTHING;

    v_persisted_sha256 :=
        platform.cutover_billing_forms_rehearsal_replay_v1(
            p_receipt_id,p_tenant_id,p_generation,
            p_billing_operation_id,p_billing_run_id,p_billing_rollback_receipt_id,
            p_billing_source_count,p_billing_target_count,
            p_billing_acceptance_sha256,p_billing_rollback_sha256,
            p_forms_operation_id,p_forms_run_id,p_forms_rollback_receipt_id,
            p_forms_source_count,p_forms_target_count,
            p_forms_acceptance_sha256,p_forms_rollback_sha256,
            p_group_sha256,p_recorded_at
        );
    IF v_persisted_sha256 IS DISTINCT FROM p_group_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='P57B4',
            MESSAGE='billing Forms rehearsal replay differs';
    END IF;
    RETURN v_persisted_sha256;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_record_billing_forms_rehearsal_v1(
    uuid,uuid,bigint,uuid,uuid,uuid,bigint,bigint,text,text,
    uuid,uuid,uuid,bigint,bigint,text,text,text,timestamptz
) FROM PUBLIC;

