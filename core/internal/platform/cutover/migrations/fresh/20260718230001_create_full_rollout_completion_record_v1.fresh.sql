-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.record_full_rollout_completion_v1(
    p_receipt_id uuid,
    p_environment_id uuid,
    p_tenant_id uuid,
    p_evidence_sha256 text,
    p_shadow_read_sha256 text,
    p_logical_delta_sha256 text,
    p_reconcile_sha256 text,
    p_production_snapshot_sha256 text,
    p_v1_removal_sha256 text,
    p_backup_restore_sha256 text,
    p_frontend_contract_sha256 text,
    p_last_legacy_hit_at timestamptz,
    p_rollback_window_closed_at timestamptz,
    p_recorded_at timestamptz
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_domain_count integer;
    v_group_count integer;
    v_member_count integer;
    v_persisted_sha256 text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_cluster_admin' OR
       p_receipt_id IS NULL OR p_environment_id IS NULL OR p_tenant_id IS NULL OR
       p_environment_id=p_tenant_id OR
       p_environment_id='00000000-0000-0000-0000-000000000000'::uuid OR
       p_tenant_id='00000000-0000-0000-0000-000000000000'::uuid OR
       NOT pg_catalog.isfinite(p_last_legacy_hit_at) OR
       NOT pg_catalog.isfinite(p_rollback_window_closed_at) OR
       NOT pg_catalog.isfinite(p_recorded_at) OR
       p_last_legacy_hit_at > p_recorded_at - interval '30 days' OR
       p_rollback_window_closed_at > p_recorded_at OR
       p_rollback_window_closed_at < p_last_legacy_hit_at OR
       p_recorded_at < pg_catalog.transaction_timestamp() - interval '5 minutes' OR
       p_recorded_at > pg_catalog.clock_timestamp() + interval '5 seconds' OR EXISTS (
        SELECT 1
        FROM pg_catalog.unnest(ARRAY[
            p_evidence_sha256,p_shadow_read_sha256,p_logical_delta_sha256,
            p_reconcile_sha256,p_production_snapshot_sha256,p_v1_removal_sha256,
            p_backup_restore_sha256,p_frontend_contract_sha256
        ]) AS supplied(value)
        WHERE supplied.value !~ '^[0-9a-f]{64}$'
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P57F0',
            MESSAGE='full rollout completion input denied';
    END IF;
    PERFORM 1
    FROM platform.domain_cutovers AS domain_row
    WHERE (domain_row.scope_kind='global' AND
           domain_row.scope_id=p_environment_id AND domain_row.domain='billing')
       OR (domain_row.scope_kind='tenant' AND
           domain_row.scope_id=p_tenant_id AND domain_row.domain IN (
               'iam','system_resource_layout','billing','workflow',
               'analytics','catalog','work','documents','notifications'
           ))
    ORDER BY domain_row.scope_kind,domain_row.domain
    FOR SHARE;
    GET DIAGNOSTICS v_domain_count = ROW_COUNT;
    IF v_domain_count IS DISTINCT FROM 10 OR EXISTS (
        SELECT 1
        FROM platform.domain_cutovers AS domain_row
        WHERE ((domain_row.scope_kind='global' AND
                domain_row.scope_id=p_environment_id AND domain_row.domain='billing')
           OR (domain_row.scope_kind='tenant' AND
               domain_row.scope_id=p_tenant_id AND domain_row.domain IN (
                   'iam','system_resource_layout','billing','workflow',
                   'analytics','catalog','work','documents','notifications'
               )))
          AND (domain_row.state IS DISTINCT FROM 'v2' OR
               domain_row.writer IS DISTINCT FROM 'v2' OR
               domain_row.approved_by IS NULL OR
               domain_row.approved_by=domain_row.requested_by OR
               domain_row.acceptance_digest IS NULL OR
               domain_row.changed_at > p_recorded_at)
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P57F1',
            MESSAGE='full rollout domain authority differs';
    END IF;
    PERFORM 1
    FROM platform.cutover_groups AS cutover_group
    WHERE cutover_group.scope_kind='tenant'
      AND cutover_group.scope_id=p_tenant_id
      AND cutover_group.group_id IN ('billing','iam_system_resources')
    ORDER BY cutover_group.group_id
    FOR SHARE;
    GET DIAGNOSTICS v_group_count = ROW_COUNT;
    IF v_group_count IS DISTINCT FROM 2 OR EXISTS (
        SELECT 1
        FROM platform.cutover_groups AS cutover_group
        WHERE cutover_group.scope_kind='tenant'
          AND cutover_group.scope_id=p_tenant_id
          AND cutover_group.group_id IN ('billing','iam_system_resources')
          AND (cutover_group.state IS DISTINCT FROM 'v2' OR
               cutover_group.writer IS DISTINCT FROM 'v2' OR
               cutover_group.approved_by IS NULL OR
               cutover_group.approved_by=cutover_group.requested_by OR
               cutover_group.acceptance_digest IS NULL OR
               cutover_group.changed_at > p_recorded_at)
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P57F2',
            MESSAGE='full rollout group authority differs';
    END IF;
    PERFORM 1
    FROM platform.cutover_group_members AS member
    WHERE member.scope_kind='tenant' AND member.scope_id=p_tenant_id
      AND member.group_id IN ('billing','iam_system_resources')
    ORDER BY member.group_id,member.ordinal
    FOR SHARE;
    GET DIAGNOSTICS v_member_count = ROW_COUNT;
    IF v_member_count IS DISTINCT FROM 3 OR
       (SELECT pg_catalog.jsonb_agg(member.domain ORDER BY member.ordinal)
        FROM platform.cutover_group_members AS member
        WHERE member.scope_kind='tenant' AND member.scope_id=p_tenant_id
          AND member.group_id='billing') IS DISTINCT FROM
          '["billing"]'::jsonb OR
       (SELECT pg_catalog.jsonb_agg(member.domain ORDER BY member.ordinal)
        FROM platform.cutover_group_members AS member
        WHERE member.scope_kind='tenant' AND member.scope_id=p_tenant_id
          AND member.group_id='iam_system_resources') IS DISTINCT FROM
          '["iam","system_resource_layout"]'::jsonb THEN
        RAISE EXCEPTION USING ERRCODE='P57F3',
            MESSAGE='full rollout group members differ';
    END IF;

    IF EXISTS (
        SELECT 1 FROM platform.contract_retirements AS retirement
        WHERE retirement.state IS DISTINCT FROM 'retired' OR
              retirement.remaining_scope_count IS DISTINCT FROM 0 OR
              retirement.zero_hit_since IS NULL OR
              retirement.zero_hit_since > p_recorded_at - interval '30 days' OR
              retirement.last_hit_at > p_recorded_at - interval '30 days'
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P57F4',
            MESSAGE='full rollout legacy retirement is incomplete';
    END IF;
    INSERT INTO cutover_authority.full_rollout_completion_receipts (
        receipt_id,environment_id,tenant_id,evidence_sha256,
        shadow_read_sha256,logical_delta_sha256,reconcile_sha256,
        production_snapshot_sha256,v1_removal_sha256,backup_restore_sha256,
        frontend_contract_sha256,
        legacy_route_hits,v1_writers,v1_consumers,v1_queue_messages,v1_credentials,
        last_legacy_hit_at,rollback_window_closed_at,recorded_at,recorded_by
    ) VALUES (
        p_receipt_id,p_environment_id,p_tenant_id,
        pg_catalog.decode(p_evidence_sha256,'hex'),
        pg_catalog.decode(p_shadow_read_sha256,'hex'),
        pg_catalog.decode(p_logical_delta_sha256,'hex'),
        pg_catalog.decode(p_reconcile_sha256,'hex'),
        pg_catalog.decode(p_production_snapshot_sha256,'hex'),
        pg_catalog.decode(p_v1_removal_sha256,'hex'),
        pg_catalog.decode(p_backup_restore_sha256,'hex'),
        pg_catalog.decode(p_frontend_contract_sha256,'hex'),
        0,0,0,0,0,p_last_legacy_hit_at,p_rollback_window_closed_at,
        p_recorded_at,SESSION_USER
    ) ON CONFLICT (receipt_id) DO NOTHING;

    SELECT pg_catalog.encode(receipt.evidence_sha256,'hex')
    INTO v_persisted_sha256
    FROM cutover_authority.full_rollout_completion_receipts AS receipt
    WHERE receipt.receipt_id=p_receipt_id
      AND receipt.environment_id=p_environment_id AND receipt.tenant_id=p_tenant_id
      AND receipt.evidence_sha256=pg_catalog.decode(p_evidence_sha256,'hex')
      AND ARRAY[receipt.shadow_read_sha256,receipt.logical_delta_sha256,
                receipt.reconcile_sha256,receipt.production_snapshot_sha256,
                receipt.v1_removal_sha256,receipt.backup_restore_sha256,
                receipt.frontend_contract_sha256] =
          ARRAY[pg_catalog.decode(p_shadow_read_sha256,'hex'),
                pg_catalog.decode(p_logical_delta_sha256,'hex'),
                pg_catalog.decode(p_reconcile_sha256,'hex'),
                pg_catalog.decode(p_production_snapshot_sha256,'hex'),
                pg_catalog.decode(p_v1_removal_sha256,'hex'),
                pg_catalog.decode(p_backup_restore_sha256,'hex'),
                pg_catalog.decode(p_frontend_contract_sha256,'hex')]
      AND receipt.last_legacy_hit_at=p_last_legacy_hit_at
      AND receipt.rollback_window_closed_at=p_rollback_window_closed_at
      AND receipt.recorded_at=p_recorded_at;
    IF v_persisted_sha256 IS DISTINCT FROM p_evidence_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='P57F5',
            MESSAGE='full rollout completion replay differs';
    END IF;
    RETURN v_persisted_sha256;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION cutover_authority.record_full_rollout_completion_v1(
    uuid,uuid,uuid,text,text,text,text,text,text,text,text,
    timestamptz,timestamptz,timestamptz
) FROM PUBLIC;
