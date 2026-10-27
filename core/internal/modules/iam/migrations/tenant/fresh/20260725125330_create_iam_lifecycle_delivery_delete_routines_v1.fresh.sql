-- +goose Up
REVOKE DELETE ON TABLE iam_lifecycle_delivery_worker_claims FROM forgecrud_runtime;
REVOKE DELETE ON TABLE iam_lifecycle_delivery_artifacts FROM forgecrud_runtime;

-- +goose StatementBegin
CREATE FUNCTION iam_lifecycle_delivery_release_claim_v1(
    p_artifact_ref text,
    p_effect_id uuid,
    p_effect_idempotency_key text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_artifact_id uuid;
BEGIN
    DELETE FROM iam_lifecycle_delivery_worker_claims
    WHERE artifact_ref = p_artifact_ref
      AND effect_id = p_effect_id
      AND effect_idempotency_key = p_effect_idempotency_key
      AND state = 'claimed'
    RETURNING artifact_id INTO v_artifact_id;
    IF v_artifact_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM lifecycle delivery claim release conflicts';
    END IF;
END;
$function$;

CREATE FUNCTION iam_lifecycle_delivery_delete_artifact_v1(
    p_artifact_id uuid,
    p_artifact_ref text,
    p_effect_id uuid,
    p_effect_idempotency_key text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_artifact_id uuid;
BEGIN
    DELETE FROM iam_lifecycle_delivery_artifacts AS artifact
    USING iam_lifecycle_delivery_worker_claims AS claim
    WHERE artifact.id = p_artifact_id
      AND artifact.artifact_ref = p_artifact_ref
      AND claim.artifact_id = artifact.id
      AND claim.effect_id = p_effect_id
      AND claim.effect_idempotency_key = p_effect_idempotency_key
      AND claim.state = 'claimed'
    RETURNING artifact.id INTO v_artifact_id;
    IF v_artifact_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM lifecycle delivery artifact deletion conflicts';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_lifecycle_delivery_release_claim_v1(text,uuid,text)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_lifecycle_delivery_release_claim_v1(text,uuid,text)
TO forgecrud_runtime;
REVOKE ALL ON FUNCTION iam_lifecycle_delivery_delete_artifact_v1(uuid,text,uuid,text)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_lifecycle_delivery_delete_artifact_v1(uuid,text,uuid,text)
TO forgecrud_runtime;
