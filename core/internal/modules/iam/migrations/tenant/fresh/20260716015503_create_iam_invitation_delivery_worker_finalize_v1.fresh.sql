-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_delivery_worker_finalize_v1(
    p_scope_schema name, p_receipt_id uuid, p_artifact_ref text,
    p_effect_id uuid, p_effect_idempotency_key text, p_message_id text,
    p_outcome text, p_code text, p_provider_reference text,
    p_result_sha256 bytea, p_completed_at timestamptz
)
RETURNS TABLE (
    receipt_id uuid, artifact_ref text, effect_idempotency_key text,
    message_id text, outcome text, code text, provider_reference text,
    result_sha256 bytea, completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_receipt_id IS NULL OR p_effect_id IS NULL OR p_completed_at IS NULL OR
       p_outcome NOT IN ('delivered','permanent_failure') OR
       p_code !~ '^[A-Z][A-Z0-9_]{1,63}$' OR
       pg_catalog.octet_length(p_result_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation delivery finalization is denied';
    END IF;
    RETURN QUERY
    WITH inserted AS (
        INSERT INTO iam_invitation_delivery_worker_receipts (
            receipt_id, artifact_id, invitation_id, token_id, artifact_ref,
            effect_id, effect_idempotency_key, message_id, outcome, code,
            provider_reference, result_sha256, completed_at
        )
        SELECT p_receipt_id, artifact.id, artifact.invitation_id, artifact.token_id,
               artifact.artifact_ref, p_effect_id, p_effect_idempotency_key,
               p_message_id, p_outcome, p_code, p_provider_reference,
               p_result_sha256, p_completed_at
        FROM iam_invitation_delivery_artifacts AS artifact
        JOIN iam_invitation_delivery_worker_claims AS claim
          ON claim.artifact_id = artifact.id
        WHERE artifact.artifact_ref = p_artifact_ref
          AND claim.effect_id = p_effect_id
          AND claim.effect_idempotency_key = p_effect_idempotency_key
          AND claim.message_id = p_message_id AND claim.state = 'claimed'
        ON CONFLICT ON CONSTRAINT iam_invitation_delivery_worker_receipts_artifact_key DO NOTHING
        RETURNING iam_invitation_delivery_worker_receipts.*
    ), consumed AS (
        DELETE FROM iam_invitation_delivery_artifacts AS artifact
        USING inserted
        WHERE artifact.id = inserted.artifact_id
        RETURNING artifact.id
    )
    SELECT inserted.receipt_id, inserted.artifact_ref,
           inserted.effect_idempotency_key, inserted.message_id,
           inserted.outcome, inserted.code, inserted.provider_reference,
           inserted.result_sha256, inserted.completed_at
    FROM inserted JOIN consumed ON consumed.id = inserted.artifact_id;
    IF NOT FOUND THEN
        RETURN QUERY
        SELECT receipt.receipt_id, receipt.artifact_ref,
               receipt.effect_idempotency_key, receipt.message_id,
               receipt.outcome, receipt.code, receipt.provider_reference,
               receipt.result_sha256, receipt.completed_at
        FROM iam_invitation_delivery_worker_receipts AS receipt
        WHERE receipt.artifact_ref = p_artifact_ref
          AND receipt.effect_id = p_effect_id
          AND receipt.effect_idempotency_key = p_effect_idempotency_key
          AND receipt.message_id = p_message_id;
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM invitation delivery finalization conflicts';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_delivery_worker_finalize_v1(
    name,uuid,text,uuid,text,text,text,text,text,bytea,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_invitation_delivery_worker_finalize_v1(
    name,uuid,text,uuid,text,text,text,text,text,bytea,timestamptz
) TO forgecrud_runtime;
