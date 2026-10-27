-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_notifications_recipient_read_v1(
    p_scope_schema name, p_recipient_id uuid
)
RETURNS TABLE (
    principal_id uuid, authorization_version bigint,
    supported_locales text[], email_target text, email_verified boolean,
    telegram_target text, telegram_verified boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_recipient_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM notification recipient read is denied';
    END IF;
    RETURN QUERY
    SELECT principal.id, principal.authorization_version,
           ARRAY['tr','en']::text[], email.normalized_value,
           email.id IS NOT NULL, NULL::text, false
    FROM iam_users AS principal
    LEFT JOIN LATERAL (
        SELECT identifier.id, identifier.normalized_value
        FROM iam_user_identifiers AS identifier
        WHERE identifier.user_id = principal.id
          AND identifier.kind = 'email_v1'
          AND identifier.state = 'active'
        ORDER BY identifier.revision DESC LIMIT 1
    ) AS email ON true
    WHERE principal.id = p_recipient_id AND principal.status = 'active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM notification recipient is unavailable';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION iam_notifications_recipient_read_v1(name,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_notifications_recipient_read_v1(name,uuid)
    TO harpy_runtime;
-- +goose StatementEnd
