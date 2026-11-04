-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_work_principal_evidence_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_target_kind text,
    in_target_id uuid
)
RETURNS TABLE (
    actor_principals bytea,
    authorization_clock bigint,
    target_exists boolean,
    target_revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::bytea,
        0::bigint,
        false,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_work_principal_evidence_v1(name, uuid, uuid, text, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_work_principal_evidence_v1(name, uuid, uuid, text, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_work_principal_evidence_v1(name, uuid, uuid, text, uuid);
