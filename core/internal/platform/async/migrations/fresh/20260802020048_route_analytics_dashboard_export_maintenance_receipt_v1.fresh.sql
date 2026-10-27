-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_route_analytics_dashboard_export_maintenance_receipt_v1(
    p_receipt platform.idempotency_receipts,p_current_user name,p_session_user name
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE v_state text;
BEGIN
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_current_user IS DISTINCT FROM 'harpy_owner' OR
       p_session_user IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_receipt.writer_origin IS DISTINCT FROM 'v2' OR
       p_receipt.effect_kind NOT IN ('analytics.dashboard.export.delete',
                                     'analytics.dashboard.export.expire') THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance receipt route is invalid';
    END IF;
    SELECT job.state INTO v_state FROM platform.jobs AS job
    WHERE job.id=p_receipt.causation_id AND job.tenant_id=p_receipt.tenant_id
    FOR UPDATE;
    IF v_state='running' THEN
        PERFORM platform.platform_async_assert_analytics_dashboard_export_running_receipt_v1(
            p_receipt,p_current_user,p_session_user);
    ELSIF v_state='dead' THEN
        PERFORM platform.platform_async_assert_analytics_dashboard_export_maintenance_receipt_v1(
            p_receipt,p_current_user,p_session_user);
    ELSE
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance receipt state is invalid';
    END IF;
END;
$function$;

CREATE FUNCTION platform.platform_async_validate_owner_insert_v10()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE v_row jsonb:=pg_catalog.to_jsonb(NEW);
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'idempotency_receipts' THEN
        RAISE EXCEPTION USING ERRCODE='P5604',
            MESSAGE='async receipt insert trigger target is invalid';
    END IF;
    IF v_row->>'effect_kind' IN ('analytics.dashboard.export.delete',
                                 'analytics.dashboard.export.expire') THEN
        PERFORM platform.platform_async_route_analytics_dashboard_export_maintenance_receipt_v1(
            NEW,CURRENT_USER,SESSION_USER);
        RETURN NEW;
    END IF;
    IF SESSION_USER='harpy_v2_runtime' AND NEW.writer_origin='v2'
       AND EXISTS (SELECT 1 FROM platform.cutover_job_bindings AS binding
                   WHERE binding.job_id=(v_row->>'causation_id')::uuid) THEN
        PERFORM platform.platform_async_assert_cutover_effect_receipt_v1(NEW);
        RETURN NEW;
    END IF;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       NEW.writer_origin IS DISTINCT FROM 'v2' THEN
        RAISE EXCEPTION USING ERRCODE='P5605',
            MESSAGE='async owner insert authority is invalid';
    END IF;
    PERFORM platform.platform_async_assert_current_authority_v1(
        NEW.scope_kind,NEW.scope_id,NEW.authority_group_id,
        NEW.authority_group_generation,NEW.authority_domain,
        NEW.cutover_generation);
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_route_analytics_dashboard_export_maintenance_receipt_v1(
        platform.idempotency_receipts,name,name) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_async_validate_owner_insert_v10()
FROM PUBLIC;
CREATE OR REPLACE TRIGGER idempotency_receipts_authority_insert_guard
BEFORE INSERT ON platform.idempotency_receipts FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v10();
ALTER TABLE platform.idempotency_receipts
    ENABLE ALWAYS TRIGGER idempotency_receipts_authority_insert_guard;
