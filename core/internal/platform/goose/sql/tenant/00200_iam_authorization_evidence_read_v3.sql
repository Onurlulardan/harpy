-- +goose Up
CREATE FUNCTION iam_authorization_evidence_read_v3(
    in_schema_name name,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_resource_key text,
    in_action_key text
)
RETURNS TABLE (
    actor_document bytea,
    resource_document bytea,
    action_document bytea,
    clock bigint,
    bindings_document bytea,
    implications_document bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::bytea,
        ''::bytea,
        ''::bytea,
        0::bigint,
        ''::bytea,
        ''::bytea
    ;
$$;

REVOKE ALL ON FUNCTION iam_authorization_evidence_read_v3(name, uuid, uuid, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_authorization_evidence_read_v3(name, uuid, uuid, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_authorization_evidence_read_v3(name, uuid, uuid, text, text);
