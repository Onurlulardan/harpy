-- +goose Up
CREATE FUNCTION iam_lifecycle_delivery_release_claim_v1(
    in_artifact_ref text,
    in_effect_id uuid,
    in_effect_key text
)
RETURNS void
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT NULL::void;
$$;

REVOKE ALL ON FUNCTION iam_lifecycle_delivery_release_claim_v1(text, uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_lifecycle_delivery_release_claim_v1(text, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_lifecycle_delivery_release_claim_v1(text, uuid, text);
