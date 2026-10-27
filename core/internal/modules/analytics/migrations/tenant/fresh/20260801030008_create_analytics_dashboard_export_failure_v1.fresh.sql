-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_outcome_unknown_v1(
    p_scope_schema name, p_export_id uuid, p_lease_owner uuid,
    p_attempt_generation bigint, p_observed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       p_lease_owner IS NULL OR
       p_lease_owner='00000000-0000-0000-0000-000000000000'::uuid OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export unknown outcome input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    UPDATE analytics_dashboard_exports
    SET failure_code='ARTIFACT_OUTCOME_UNKNOWN',lease_owner=NULL,
        lease_until=NULL,row_revision=row_revision+1,updated_at=p_observed_at
    WHERE export_id=p_export_id AND state='running'
      AND lease_owner=p_lease_owner
      AND attempt_generation=p_attempt_generation
      AND effect_started_at IS NOT NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export unknown outcome changed';
    END IF;
    RETURN true;
END;
$function$;

CREATE FUNCTION analytics_dashboard_export_fail_render_v1(
    p_scope_schema name, p_export_id uuid, p_lease_owner uuid,
    p_attempt_generation bigint, p_failure_code text,
    p_retryable boolean, p_failed_at timestamptz
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_next text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       p_lease_owner IS NULL OR
       p_lease_owner='00000000-0000-0000-0000-000000000000'::uuid OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.octet_length(p_failure_code) NOT BETWEEN 3 AND 64 OR
       p_retryable IS NULL OR p_failed_at IS NULL OR
       NOT pg_catalog.isfinite(p_failed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export failure input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    v_next:=CASE WHEN p_retryable AND p_attempt_generation<3
                 THEN 'queued' ELSE 'failed' END;
    UPDATE analytics_dashboard_exports
    SET state=v_next,
        failure_code=CASE WHEN v_next='failed' THEN p_failure_code ELSE NULL END,
        lease_owner=NULL,lease_until=NULL,row_revision=row_revision+1,
        updated_at=p_failed_at
    WHERE export_id=p_export_id AND state='running'
      AND lease_owner=p_lease_owner
      AND attempt_generation=p_attempt_generation
      AND effect_started_at IS NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export failure changed';
    END IF;
    RETURN v_next;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_outcome_unknown_v1(
    name,uuid,uuid,bigint,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_fail_render_v1(
    name,uuid,uuid,bigint,text,boolean,timestamptz
) FROM PUBLIC;
