-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_delivery_worker_claim_v1(
    p_scope_schema name, p_artifact_ref text, p_effect_id uuid,
    p_effect_idempotency_key text, p_message_id text, p_claimed_at timestamptz
)
RETURNS TABLE (
    artifact_id uuid, invitation_id uuid, token_id uuid, artifact_ref text,
    recipient text, kms_key_id text, nonce bytea, ciphertext bytea,
    authentication_tag bytea, aad_sha256 bytea, ciphertext_sha256 bytea,
    expires_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_effect_id IS NULL OR p_claimed_at IS NULL OR
       pg_catalog.char_length(p_effect_idempotency_key) NOT BETWEEN 1 AND 512 OR
       pg_catalog.char_length(p_message_id) NOT BETWEEN 3 AND 512 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation delivery claim is denied';
    END IF;
    RETURN QUERY
    WITH candidate AS MATERIALIZED (
        SELECT artifact.id, artifact.invitation_id, artifact.token_id,
               artifact.artifact_ref, invitation.normalized_value,
               artifact.kms_key_id, artifact.nonce, artifact.ciphertext,
               artifact.authentication_tag, artifact.aad_sha256,
               artifact.ciphertext_sha256, artifact.expires_at
        FROM iam_invitation_delivery_artifacts AS artifact
        JOIN iam_invitations AS invitation ON invitation.id = artifact.invitation_id
        WHERE artifact.artifact_ref = p_artifact_ref
          AND invitation.state = 'pending'
          AND invitation.expires_at > pg_catalog.statement_timestamp()
          AND artifact.expires_at > pg_catalog.statement_timestamp()
        FOR UPDATE OF artifact
    ), claimed AS (
        INSERT INTO iam_invitation_delivery_worker_claims (
            artifact_id, artifact_ref, effect_id, effect_idempotency_key,
            message_id, state, claimed_at, outcome_unknown_at
        )
        SELECT candidate.id, candidate.artifact_ref, p_effect_id,
               p_effect_idempotency_key, p_message_id, 'claimed', p_claimed_at, NULL
        FROM candidate
        ON CONFLICT ON CONSTRAINT iam_invitation_delivery_worker_claims_pkey DO NOTHING
        RETURNING iam_invitation_delivery_worker_claims.artifact_id
    )
    SELECT candidate.id, candidate.invitation_id, candidate.token_id,
           candidate.artifact_ref, candidate.normalized_value,
           candidate.kms_key_id, candidate.nonce, candidate.ciphertext,
           candidate.authentication_tag, candidate.aad_sha256,
           candidate.ciphertext_sha256, candidate.expires_at
    FROM candidate JOIN claimed ON claimed.artifact_id = candidate.id;
END;
$function$;

CREATE FUNCTION iam_invitation_delivery_worker_release_v1(
    p_scope_schema name, p_artifact_ref text, p_effect_id uuid,
    p_effect_idempotency_key text
)
RETURNS SETOF boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_effect_id IS NULL OR pg_catalog.char_length(p_artifact_ref) < 1 OR
       pg_catalog.char_length(p_effect_idempotency_key) NOT BETWEEN 1 AND 512 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation delivery claim release is denied';
    END IF;
    RETURN QUERY
    WITH released AS (
        DELETE FROM iam_invitation_delivery_worker_claims AS claim
        WHERE claim.artifact_ref = p_artifact_ref
          AND claim.effect_id = p_effect_id
          AND claim.effect_idempotency_key = p_effect_idempotency_key
          AND claim.state = 'claimed'
        RETURNING true AS released
    )
    SELECT released.released FROM released;
END;
$function$;

CREATE FUNCTION iam_invitation_delivery_worker_mark_unknown_v1(
    p_scope_schema name, p_artifact_ref text, p_effect_id uuid,
    p_effect_idempotency_key text, p_outcome_unknown_at timestamptz
)
RETURNS SETOF boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_effect_id IS NULL OR p_outcome_unknown_at IS NULL OR
       pg_catalog.char_length(p_artifact_ref) < 1 OR
       pg_catalog.char_length(p_effect_idempotency_key) NOT BETWEEN 1 AND 512 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation delivery unknown outcome is denied';
    END IF;
    RETURN QUERY
    WITH marked AS (
        UPDATE iam_invitation_delivery_worker_claims AS claim
        SET state = 'outcome_unknown',
            outcome_unknown_at = p_outcome_unknown_at
        WHERE claim.artifact_ref = p_artifact_ref
          AND claim.effect_id = p_effect_id
          AND claim.effect_idempotency_key = p_effect_idempotency_key
          AND (claim.state = 'claimed' OR
               (claim.state = 'outcome_unknown' AND
                claim.outcome_unknown_at = p_outcome_unknown_at))
        RETURNING true AS marked
    )
    SELECT marked.marked FROM marked;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_delivery_worker_claim_v1(
    name,text,uuid,text,text,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_invitation_delivery_worker_claim_v1(
    name,text,uuid,text,text,timestamptz
) TO forgecrud_runtime;
REVOKE ALL ON FUNCTION iam_invitation_delivery_worker_release_v1(
    name,text,uuid,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_invitation_delivery_worker_release_v1(
    name,text,uuid,text
) TO forgecrud_runtime;
REVOKE ALL ON FUNCTION iam_invitation_delivery_worker_mark_unknown_v1(
    name,text,uuid,text,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_invitation_delivery_worker_mark_unknown_v1(
    name,text,uuid,text,timestamptz
) TO forgecrud_runtime;
