-- +goose Up
CREATE FUNCTION platform.platformauthority_fence_signer_v1(
    in_binding_id text,
    in_expected_signer_generation bigint,
    in_expected_signer_kid text,
    in_evidence_sha256 text,
    in_now timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformauthority_fence_signer_v1(text, bigint, text, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_fence_signer_v1(text, bigint, text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_fence_signer_v1(text, bigint, text, text, timestamptz);
