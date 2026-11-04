-- +goose Up
CREATE FUNCTION iam_invitation_delivery_worker_finalize_v1(
    in_schema_name name,
    in_receipt_id uuid,
    in_artifact_ref text,
    in_effect_id uuid,
    in_effect_key text,
    in_message_id text,
    in_outcome text,
    in_code text,
    in_provider_reference text,
    in_result_sha256 bytea,
    in_completed_at timestamptz
)
RETURNS TABLE (
    receipt_id uuid,
    artifact_ref text,
    effect_idempotency_key text,
    message_id text,
    outcome text,
    code text,
    provider_reference text,
    result_sha256 bytea,
    completed_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_receipt_id,
        in_artifact_ref,
        ''::text,
        in_message_id,
        in_outcome,
        in_code,
        in_provider_reference,
        in_result_sha256,
        in_completed_at
    ;
$$;

REVOKE ALL ON FUNCTION iam_invitation_delivery_worker_finalize_v1(name, uuid, text, uuid, text, text, text, text, text, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_invitation_delivery_worker_finalize_v1(name, uuid, text, uuid, text, text, text, text, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_invitation_delivery_worker_finalize_v1(name, uuid, text, uuid, text, text, text, text, text, bytea, timestamptz);
