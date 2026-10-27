-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_one_time_issue_v1(
    p_token_id uuid, p_purpose text, p_user_id uuid, p_identifier_id uuid,
    p_identifier_kind text, p_identifier_revision bigint,
    p_portal_binding_id uuid, p_portal_setup_revision bigint,
    p_selector bytea, p_hash_key_id text, p_secret_hmac bytea,
    p_binding_sha256 bytea
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_expires timestamptz;
BEGIN
    IF p_token_id IS NULL OR p_user_id IS NULL OR p_identifier_id IS NULL OR
       p_identifier_kind NOT IN ('email_v1', 'portal_v1') OR
       p_identifier_revision < 1 OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       p_hash_key_id !~ '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       pg_catalog.octet_length(p_secret_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_binding_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM one-time token issue input is invalid';
    END IF;
    CASE p_purpose
      WHEN 'email_verification_v1' THEN
        IF p_identifier_kind <> 'email_v1' OR p_portal_binding_id IS NOT NULL OR
           p_portal_setup_revision IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '22023';
        END IF;
        v_expires := v_now + INTERVAL '24 hours';
      WHEN 'password_reset_v1' THEN
        IF p_portal_binding_id IS NOT NULL OR
           p_portal_setup_revision IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '22023';
        END IF;
        v_expires := v_now + INTERVAL '1 hour';
      WHEN 'portal_setup_v1' THEN
        IF p_identifier_kind <> 'portal_v1' OR p_portal_binding_id IS NULL OR
           p_portal_setup_revision < 1 THEN
            RAISE EXCEPTION USING ERRCODE = '22023';
        END IF;
        v_expires := v_now + INTERVAL '1 hour';
      ELSE RAISE EXCEPTION USING ERRCODE = '22023';
    END CASE;
    UPDATE iam_one_time_tokens SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE purpose = p_purpose AND user_id = p_user_id
      AND identifier_id = p_identifier_id AND state = 'pending';
    INSERT INTO iam_one_time_tokens (
        id, purpose, user_id, identifier_id, identifier_kind,
        identifier_revision, portal_binding_id, portal_setup_revision,
        selector, hash_key_id, secret_hmac, binding_sha256, state,
        attempt_count, revision, created_at, updated_at, expires_at, purge_after
    ) VALUES (
        p_token_id, p_purpose, p_user_id, p_identifier_id, p_identifier_kind,
        p_identifier_revision, p_portal_binding_id, p_portal_setup_revision,
        p_selector, p_hash_key_id, p_secret_hmac, p_binding_sha256, 'pending',
        0, 1, v_now, v_now, v_expires, v_expires + INTERVAL '24 hours'
    );
    RETURN 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_one_time_issue_v1(
    uuid,text,uuid,uuid,text,bigint,uuid,bigint,bytea,text,bytea,bytea
) FROM PUBLIC;
