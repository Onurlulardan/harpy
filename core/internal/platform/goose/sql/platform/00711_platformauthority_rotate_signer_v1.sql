-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_rotate_signer_v1(
    in_operation_id text,
    in_binding_id text,
    in_expected_signer_generation bigint,
    in_expected_signer_kid text,
    in_new_signer_generation bigint,
    in_new_signer_kid text,
    in_new_signer_public_key_sha256 text,
    in_new_recovery_trust_bundle_sha256 text,
    in_now timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_rotate_signer_v1(text, text, bigint, text, bigint, text, text, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_rotate_signer_v1(text, text, bigint, text, bigint, text, text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_rotate_signer_v1(text, text, bigint, text, bigint, text, text, text, timestamptz);
