-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_delivery_worker_claim_v1(
    in_schema_name name,
    in_artifact_ref text,
    in_effect_id uuid,
    in_effect_key text,
    in_message_id text,
    in_claimed_at timestamptz
)
RETURNS TABLE (
    artifact_id uuid,
    invitation_id uuid,
    token_id uuid,
    artifact_ref text,
    recipient text,
    kms_key_id text,
    nonce bytea,
    ciphertext bytea,
    authentication_tag bytea,
    aad_sha256 bytea,
    ciphertext_sha256 bytea,
    expires_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        in_artifact_ref,
        ''::text,
        ''::text,
        ''::bytea,
        ''::bytea,
        ''::bytea,
        ''::bytea,
        ''::bytea,
        pg_catalog.clock_timestamp()
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_delivery_worker_claim_v1(name, text, uuid, text, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_invitation_delivery_worker_claim_v1(name, text, uuid, text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_invitation_delivery_worker_claim_v1(name, text, uuid, text, text, timestamptz);
