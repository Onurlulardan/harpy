-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_user_bump_v1(
    p_user_id uuid, p_bump_token boolean
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_revision bigint;
BEGIN
    UPDATE iam_users
    SET authorization_version = authorization_version + 1,
        token_version = token_version + CASE WHEN p_bump_token THEN 1 ELSE 0 END,
        revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
    WHERE id = p_user_id AND status <> 'tombstoned'
    RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM identity user is unavailable';
    END IF;
    RETURN v_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_user_bump_v1(uuid,boolean) FROM PUBLIC;
