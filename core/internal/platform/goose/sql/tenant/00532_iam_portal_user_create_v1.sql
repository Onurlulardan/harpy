-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_portal_user_create_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_user_id uuid,
    in_schema_id uuid,
    in_values bytea,
    in_profile_sha bytea,
    in_id uuid,
    in_display text,
    in_normalized text,
    in_normalized_sha256 bytea,
    in_id_2 uuid,
    in_display_2 text,
    in_normalized_2 text,
    in_normalized_sha256_2 bytea,
    in_access_mode_id uuid,
    in_binding_id uuid,
    in_type text,
    in_form_id uuid,
    in_form_version_id uuid,
    in_record_id uuid,
    in_identifier_field_id uuid,
    in_contract_sha256 bytea,
    in_setup_state text,
    in_role_names text[],
    in_role_assignment_i_ds uuid[],
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid,
    revision bigint,
    authorization_clock bigint,
    replayed boolean,
    email_identifier_id uuid,
    email_revision bigint,
    portal_identifier_id uuid,
    portal_revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_id,
        0::bigint,
        0::bigint,
        false,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_user_create_v1(name, uuid, uuid, bytea, uuid, uuid, bytea, bytea, uuid, text, text, bytea, uuid, text, text, bytea, uuid, uuid, text, uuid, uuid, uuid, uuid, bytea, text, text[], uuid[], uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_portal_user_create_v1(name, uuid, uuid, bytea, uuid, uuid, bytea, bytea, uuid, text, text, bytea, uuid, text, text, bytea, uuid, uuid, text, uuid, uuid, uuid, uuid, bytea, text, text[], uuid[], uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_portal_user_create_v1(name, uuid, uuid, bytea, uuid, uuid, bytea, bytea, uuid, text, text, bytea, uuid, text, text, bytea, uuid, uuid, text, uuid, uuid, uuid, uuid, bytea, text, text[], uuid[], uuid, uuid, bigint, bigint);
