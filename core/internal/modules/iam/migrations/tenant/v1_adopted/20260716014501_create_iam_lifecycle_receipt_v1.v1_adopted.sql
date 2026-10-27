-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_lifecycle_replay_v1(
    p_operation_id uuid, p_command_key text, p_request_sha256 bytea,
    p_result_id uuid
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_command_key text;
    v_request_sha256 bytea;
    v_result_id uuid;
    v_subject_user_id uuid;
    v_result_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
BEGIN
    IF p_operation_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM lifecycle replay input is invalid';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            pg_catalog.current_schema() || ':' || p_operation_id::text, 0
        )
    );
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_id,
        receipt.subject_user_id, receipt.result_revision,
        receipt.token_version, receipt.authorization_version
    INTO v_command_key, v_request_sha256, v_result_id, v_subject_user_id,
        v_result_revision, v_token_version, v_authorization_version
    FROM iam_lifecycle_operation_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id;
    IF NOT FOUND THEN RETURN; END IF;
    IF v_command_key IS DISTINCT FROM p_command_key OR
       v_request_sha256 IS DISTINCT FROM p_request_sha256 OR
       (p_result_id IS NOT NULL AND
        v_result_id IS DISTINCT FROM p_result_id) THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'IAM lifecycle operation identity was reused';
    END IF;
    RETURN QUERY SELECT v_result_id, v_subject_user_id, v_result_revision,
        v_token_version, v_authorization_version, true;
END;
$function$;

CREATE FUNCTION iam_lifecycle_record_v1(
    p_operation_id uuid, p_command_key text, p_request_sha256 bytea,
    p_result_id uuid, p_subject_user_id uuid, p_result_revision bigint,
    p_source_kind text, p_actor_id uuid, p_resource_kind text,
    p_bump_token boolean, p_bump_authorization boolean
)
RETURNS TABLE (token_version bigint, authorization_version bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_operation_id IS NULL OR p_result_id IS NULL OR
       p_subject_user_id IS NULL OR p_actor_id IS NULL OR
       p_result_revision < 1 OR p_source_kind NOT IN (
           'actor_tenant', 'credential_target'
       ) OR p_resource_kind !~ '^[a-z][a-z0-9_.]{1,63}$' OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM lifecycle record input is invalid';
    END IF;
    UPDATE iam_users AS subject
    SET token_version = subject.token_version + p_bump_token::integer,
        authorization_version = subject.authorization_version +
            p_bump_authorization::integer,
        revision = subject.revision + 1, updated_at = v_now
    WHERE subject.id = p_subject_user_id
      AND subject.status <> 'tombstoned'
    RETURNING subject.token_version, subject.authorization_version
    INTO token_version, authorization_version;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '23503',
            MESSAGE = 'IAM lifecycle subject is unavailable';
    END IF;
    INSERT INTO iam_lifecycle_operation_receipts (
        operation_id, command_key, request_sha256, result_id,
        subject_user_id, result_revision, token_version,
        authorization_version, created_at
    ) VALUES (
        p_operation_id, p_command_key, p_request_sha256, p_result_id,
        p_subject_user_id, p_result_revision, token_version,
        authorization_version, v_now
    );
    INSERT INTO iam_lifecycle_audit (
        operation_id, command_key, source_kind, actor_id, subject_user_id,
        resource_id, resource_kind, resource_revision, token_version,
        authorization_version, request_sha256, occurred_at
    ) VALUES (
        p_operation_id, p_command_key, p_source_kind, p_actor_id,
        p_subject_user_id, p_result_id, p_resource_kind, p_result_revision,
        token_version, authorization_version, p_request_sha256, v_now
    );
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_lifecycle_replay_v1(uuid,text,bytea,uuid)
    FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_lifecycle_record_v1(
    uuid,text,bytea,uuid,uuid,bigint,text,uuid,text,boolean,boolean
) FROM PUBLIC;
