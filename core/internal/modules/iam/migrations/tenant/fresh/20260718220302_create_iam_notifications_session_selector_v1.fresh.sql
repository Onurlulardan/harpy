-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_notifications_session_selector_v1(
    p_scope_schema name, p_principal_id uuid,
    p_session_id uuid, p_now timestamptz
)
RETURNS bytea
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_selector bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_principal_id IS NULL OR p_session_id IS NULL OR p_now IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM notification session read is denied';
    END IF;
    SELECT token.selector
    INTO STRICT v_selector
    FROM iam_sessions AS session
    JOIN iam_refresh_families AS family
      ON family.session_id = session.id AND family.user_id = session.user_id
    JOIN iam_refresh_tokens AS token ON token.family_id = family.id
    WHERE session.id = p_session_id
      AND session.user_id = p_principal_id
      AND session.access_mode = 'operator_console'
      AND session.state = 'active'
      AND family.state = 'active'
      AND token.state = 'active'
      AND session.absolute_expires_at > p_now
      AND family.absolute_expires_at > p_now
      AND token.idle_expires_at > p_now;
    IF pg_catalog.octet_length(v_selector) IS DISTINCT FROM 16 THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM notification session is unavailable';
    END IF;
    RETURN v_selector;
EXCEPTION
    WHEN NO_DATA_FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM notification session is unavailable';
END;
$function$;

REVOKE ALL ON FUNCTION iam_notifications_session_selector_v1(
    name,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_notifications_session_selector_v1(
    name,uuid,uuid,timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
