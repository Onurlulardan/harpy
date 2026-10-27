-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_session_list_v1(
    p_scope_schema name, p_user_id uuid, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint,
    p_limit integer, p_after_last_seen_at timestamptz, p_after_id uuid
)
RETURNS TABLE (
    session_id uuid, access_mode text, portal_binding_id uuid,
    device_kind text, browser_family text, os_family text,
    revision bigint, created_at timestamptz, last_seen_at timestamptz,
    absolute_expires_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_user_id IS NULL OR p_user_id <> p_actor_id OR
       p_limit NOT BETWEEN 1 AND 100 OR
       ((p_after_last_seen_at IS NULL) <> (p_after_id IS NULL)) OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY
    SELECT session.id, session.access_mode, session.portal_binding_id,
        session.device_kind, session.browser_family, session.os_family,
        session.revision, session.created_at, session.last_seen_at,
        session.absolute_expires_at
    FROM iam_sessions AS session
    WHERE session.user_id = p_user_id AND session.state = 'active'
      AND session.absolute_expires_at > pg_catalog.transaction_timestamp()
      AND (p_after_last_seen_at IS NULL OR
          (session.last_seen_at, session.id) <
          (p_after_last_seen_at, p_after_id))
    ORDER BY session.last_seen_at DESC, session.id DESC
    LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_session_list_v1(
    name,uuid,uuid,uuid,bigint,bigint,integer,timestamptz,uuid
) FROM PUBLIC;

