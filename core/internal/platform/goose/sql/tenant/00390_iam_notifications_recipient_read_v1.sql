-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_notifications_recipient_read_v1(
    arg_1 name,
    arg_2 uuid
)
RETURNS TABLE (
    principal_id uuid,
    authorization_version bigint,
    supported_locales text[],
    email_target text,
    email_verified boolean,
    telegram_target text,
    telegram_verified boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        NULL::text[],
        NULL::text,
        false,
        NULL::text,
        false;
$$;

REVOKE ALL ON FUNCTION iam_notifications_recipient_read_v1(name, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_notifications_recipient_read_v1(name, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_notifications_recipient_read_v1(name, uuid);
