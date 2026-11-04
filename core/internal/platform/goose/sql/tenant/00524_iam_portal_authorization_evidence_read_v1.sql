-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_portal_authorization_evidence_read_v1(
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
    authorization_clock bigint,
    bindings_document bytea,
    implications_document bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::bytea,
        ''::bytea,
        ''::bytea,
        0::bigint,
        ''::bytea,
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_authorization_evidence_read_v1(name, uuid, uuid, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_portal_authorization_evidence_read_v1(name, uuid, uuid, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_portal_authorization_evidence_read_v1(name, uuid, uuid, text, text);
