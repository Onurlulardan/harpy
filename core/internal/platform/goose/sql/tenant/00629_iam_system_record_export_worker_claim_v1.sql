-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_worker_claim_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_export_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_lease_owner_id uuid,
    in_attempt_generation bigint,
    in_resource text,
    in_token_version bigint,
    in_authorization_version bigint
)
RETURNS TABLE (
    state text,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_worker_claim_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_claim_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_export_worker_claim_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint);
