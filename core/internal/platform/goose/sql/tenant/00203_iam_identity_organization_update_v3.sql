-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_organization_update_v3(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_id uuid,
    in_expected_revision bigint,
    in_name text,
    in_slug text,
    in_parent_id uuid,
    in_replace_owners boolean,
    in_owner_ids uuid[],
    in_owner_membership_ids uuid[],
    in_schema_id uuid,
    in_values bytea,
    in_digest bytea,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid,
    result_revision bigint,
    authorization_clock bigint,
    replayed boolean
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
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_organization_update_v3(name, uuid, bytea, uuid, bigint, text, text, uuid, boolean, uuid[], uuid[], uuid, bytea, bytea, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_identity_organization_update_v3(name, uuid, bytea, uuid, bigint, text, text, uuid, boolean, uuid[], uuid[], uuid, bytea, bytea, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_identity_organization_update_v3(name, uuid, bytea, uuid, bigint, text, text, uuid, boolean, uuid[], uuid[], uuid, bytea, bytea, uuid, uuid, bigint, bigint);
