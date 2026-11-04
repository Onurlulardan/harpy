-- +goose Up
CREATE FUNCTION iam_password_reset_delivery_target_v1(
    in_schema_name name,
    in_identifier_kind text,
    in_identifier_sha256 bytea
)
RETURNS TABLE (
    subject_user_id uuid,
    resource_id uuid,
    recipient_identifier_id uuid,
    recipient_address text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text
    ;
$$;

REVOKE ALL ON FUNCTION iam_password_reset_delivery_target_v1(name, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_password_reset_delivery_target_v1(name, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_password_reset_delivery_target_v1(name, text, bytea);
