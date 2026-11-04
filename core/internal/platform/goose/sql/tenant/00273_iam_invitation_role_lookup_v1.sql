-- +goose Up
CREATE FUNCTION iam_invitation_role_lookup_v1(
    in_schema_name name,
    in_actor_id uuid,
    in_name text
)
RETURNS TABLE (
    id uuid,
    name text,
    kind text,
    id_2 uuid,
    organization_id uuid
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        in_name,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        NULL::uuid
    ;
$$;

REVOKE ALL ON FUNCTION iam_invitation_role_lookup_v1(name, uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_invitation_role_lookup_v1(name, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_invitation_role_lookup_v1(name, uuid, text);
