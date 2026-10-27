-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.support_list_bugs_v1(
    p_actor_id uuid, p_authorization_version bigint, p_authority_clock bigint,
    p_platform_scope boolean, p_tenant_id uuid, p_reporter_id uuid,
    p_status text, p_severity text, p_category text, p_page integer, p_page_size integer
) RETURNS TABLE(items jsonb, total bigint)
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF p_actor_id IS NULL OR p_authorization_version < 1 OR p_authority_clock < 1 OR
       p_platform_scope IS NULL OR (NOT p_platform_scope AND p_tenant_id IS NULL) OR
       (p_platform_scope AND p_reporter_id IS NOT NULL) OR
       (p_status IS NOT NULL AND p_status NOT IN ('open','in_progress','resolved','closed','wont_fix','duplicate')) OR
       (p_severity IS NOT NULL AND p_severity NOT IN ('critical','high','medium','low')) OR
       (p_category IS NOT NULL AND p_category NOT IN ('bug','feature_request','performance','ui_issue','data_issue','other')) OR
       p_page NOT BETWEEN 1 AND 5001 OR p_page_size NOT BETWEEN 1 AND 100 OR
       (p_page - 1) * p_page_size > 5000 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6501', MESSAGE = 'support bug list invalid';
    END IF;
    RETURN QUERY
    WITH filtered AS MATERIALIZED (
        SELECT report.* FROM platform.support_bug_reports AS report
        WHERE ((p_platform_scope AND (p_tenant_id IS NULL OR report.tenant_id = p_tenant_id))
            OR (NOT p_platform_scope AND report.tenant_id = p_tenant_id))
          AND (p_reporter_id IS NULL OR report.reporter_id = p_reporter_id)
          AND (p_status IS NULL OR report.status = p_status)
          AND (p_severity IS NULL OR report.severity = p_severity)
          AND (p_category IS NULL OR report.category = p_category)
    ), page_rows AS (
        SELECT filtered.* FROM filtered
        ORDER BY filtered.created_at DESC, filtered.id DESC
        OFFSET ((p_page - 1) * p_page_size) LIMIT p_page_size
    )
    SELECT COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(row_value)
        ORDER BY row_value.created_at DESC, row_value.id DESC) FROM page_rows AS row_value),
        '[]'::jsonb), (SELECT pg_catalog.count(*) FROM filtered);
END;
$function$;

CREATE FUNCTION platform.support_bug_stats_v1(
    p_actor_id uuid, p_authorization_version bigint, p_authority_clock bigint,
    p_platform_scope boolean, p_tenant_id uuid
) RETURNS jsonb
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_stats jsonb;
BEGIN
    IF p_actor_id IS NULL OR p_authorization_version < 1 OR p_authority_clock < 1 OR
       p_platform_scope IS NULL OR (NOT p_platform_scope AND p_tenant_id IS NULL) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6501', MESSAGE = 'support bug stats invalid';
    END IF;
    WITH visible AS (
        SELECT report.* FROM platform.support_bug_reports AS report
        WHERE p_platform_scope OR report.tenant_id = p_tenant_id
    ), categories AS (
        SELECT COALESCE(pg_catalog.jsonb_object_agg(value.category, value.amount), '{}'::jsonb) AS document
        FROM (SELECT visible.category, pg_catalog.count(*) AS amount FROM visible GROUP BY visible.category) AS value
    ), severities AS (
        SELECT COALESCE(pg_catalog.jsonb_object_agg(value.severity, value.amount), '{}'::jsonb) AS document
        FROM (SELECT visible.severity, pg_catalog.count(*) AS amount FROM visible GROUP BY visible.severity) AS value
    )
    SELECT pg_catalog.jsonb_build_object(
        'total', pg_catalog.count(*),
        'open', pg_catalog.count(*) FILTER (WHERE visible.status = 'open'),
        'in_progress', pg_catalog.count(*) FILTER (WHERE visible.status = 'in_progress'),
        'resolved', pg_catalog.count(*) FILTER (WHERE visible.status = 'resolved'),
        'closed', pg_catalog.count(*) FILTER (WHERE visible.status = 'closed'),
        'critical_open', pg_catalog.count(*) FILTER (WHERE visible.status = 'open' AND visible.severity = 'critical'),
        'by_category', (SELECT categories.document FROM categories),
        'by_severity', (SELECT severities.document FROM severities)
    ) INTO v_stats FROM visible;
    RETURN v_stats;
END;
$function$;

CREATE FUNCTION platform.support_get_bug_v1(
    p_actor_id uuid, p_authorization_version bigint, p_authority_clock bigint,
    p_tenant_id uuid, p_reporter_id uuid, p_id uuid
) RETURNS TABLE(response_document jsonb)
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF p_actor_id IS NULL OR p_authorization_version < 1 OR p_authority_clock < 1 OR
       p_tenant_id IS NULL OR p_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P6501', MESSAGE = 'support bug read invalid';
    END IF;
    RETURN QUERY SELECT pg_catalog.to_jsonb(report)
    FROM platform.support_bug_reports AS report
    WHERE report.id = p_id AND report.tenant_id = p_tenant_id
      AND (p_reporter_id IS NULL OR report.reporter_id = p_reporter_id);
END;
$function$;

REVOKE ALL ON FUNCTION platform.support_list_bugs_v1(uuid,bigint,bigint,boolean,uuid,uuid,text,text,text,integer,integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.support_bug_stats_v1(uuid,bigint,bigint,boolean,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.support_get_bug_v1(uuid,bigint,bigint,uuid,uuid,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_list_bugs_v1(uuid,bigint,bigint,boolean,uuid,uuid,text,text,text,integer,integer) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.support_bug_stats_v1(uuid,bigint,bigint,boolean,uuid) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.support_get_bug_v1(uuid,bigint,bigint,uuid,uuid,uuid) TO forgecrud_runtime;
-- +goose StatementEnd
