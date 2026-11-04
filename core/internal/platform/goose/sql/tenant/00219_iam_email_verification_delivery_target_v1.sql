-- +goose Up
CREATE FUNCTION iam_email_verification_delivery_target_v1(
    in_schema_name name,
    in_identifier_id uuid,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
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

REVOKE ALL ON FUNCTION iam_email_verification_delivery_target_v1(name, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_email_verification_delivery_target_v1(name, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_email_verification_delivery_target_v1(name, uuid, uuid, uuid, bigint, bigint);
