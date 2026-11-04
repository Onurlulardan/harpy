-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_claim_recovery_manifest_v2(
    in_artifact_id text,
    in_tenant_id text,
    in_actor_id text,
    in_session_id text
)
RETURNS TABLE (
    sha256 text,
    bytes bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        ''::bytea;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_claim_recovery_manifest_v2(text, text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_claim_recovery_manifest_v2(text, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_claim_recovery_manifest_v2(text, text, text, text);
