-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.support_list_contacts_v1(
    p_actor_id uuid, p_authorization_version bigint, p_authority_clock bigint,
    p_status text, p_contact_type text, p_page integer, p_page_size integer
) RETURNS TABLE(items jsonb, total bigint)
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF p_actor_id IS NULL OR p_authorization_version < 1 OR p_authority_clock < 1 OR
       (p_status IS NOT NULL AND p_status NOT IN ('new','read','replied','archived')) OR
       (p_contact_type IS NOT NULL AND p_contact_type NOT IN ('general','demo_request','pricing_inquiry','technical_question','partnership','enterprise_plan','on_premise')) OR
       p_page NOT BETWEEN 1 AND 5001 OR p_page_size NOT BETWEEN 1 AND 100 OR
       (p_page - 1) * p_page_size > 5000 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6501', MESSAGE = 'support contact list invalid';
    END IF;
    RETURN QUERY
    WITH filtered AS MATERIALIZED (
        SELECT message.* FROM platform.support_contact_messages AS message
        WHERE (p_status IS NULL OR message.status = p_status)
          AND (p_contact_type IS NULL OR message.contact_type = p_contact_type)
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

CREATE FUNCTION platform.support_contact_stats_v1(
    p_actor_id uuid, p_authorization_version bigint, p_authority_clock bigint,
    p_observed_at timestamp(6) with time zone
) RETURNS jsonb
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_stats jsonb;
BEGIN
    IF p_actor_id IS NULL OR p_authorization_version < 1 OR p_authority_clock < 1 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) OR
       pg_catalog.date_trunc('microseconds', p_observed_at) <> p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6501', MESSAGE = 'support contact stats invalid';
    END IF;
    SELECT pg_catalog.jsonb_build_object(
        'total', pg_catalog.count(*),
        'new', pg_catalog.count(*) FILTER (WHERE message.status = 'new'),
        'read', pg_catalog.count(*) FILTER (WHERE message.status = 'read'),
        'replied', pg_catalog.count(*) FILTER (WHERE message.status = 'replied'),
        'archived', pg_catalog.count(*) FILTER (WHERE message.status = 'archived'),
        'this_month', pg_catalog.count(*) FILTER (WHERE message.created_at >=
            pg_catalog.date_trunc('month', p_observed_at, 'UTC'))
    ) INTO v_stats FROM platform.support_contact_messages AS message;
    RETURN v_stats;
END;
$function$;

CREATE FUNCTION platform.support_get_contact_v1(
    p_actor_id uuid, p_authorization_version bigint, p_authority_clock bigint, p_id uuid
) RETURNS TABLE(response_document jsonb)
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF p_actor_id IS NULL OR p_authorization_version < 1 OR p_authority_clock < 1 OR p_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P6501', MESSAGE = 'support contact read invalid';
    END IF;
    RETURN QUERY SELECT pg_catalog.to_jsonb(message)
    FROM platform.support_contact_messages AS message WHERE message.id = p_id;
END;
$function$;

REVOKE ALL ON FUNCTION platform.support_list_contacts_v1(uuid,bigint,bigint,text,text,integer,integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.support_contact_stats_v1(uuid,bigint,bigint,timestamp with time zone) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.support_get_contact_v1(uuid,bigint,bigint,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_list_contacts_v1(uuid,bigint,bigint,text,text,integer,integer) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.support_contact_stats_v1(uuid,bigint,bigint,timestamp with time zone) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.support_get_contact_v1(uuid,bigint,bigint,uuid) TO forgecrud_runtime;
-- +goose StatementEnd
