-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_expire_v2(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_invitation_id uuid, p_expected_revision bigint, p_actor_user_id uuid,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    invitation_id uuid, invitation_state text,
    invitation_revision bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_id uuid;
    v_state text;
    v_revision bigint;
    v_command text;
    v_actor uuid;
    v_request bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_operation_id IS NULL OR p_invitation_id IS NULL OR
       p_actor_user_id IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation expiry scope is invalid';
    END IF;
    SELECT command_key, actor_id, request_sha256, result_id,
           result_state, result_revision
    INTO v_command, v_actor, v_request, v_id, v_state, v_revision
    FROM iam_invitation_operation_receipts WHERE operation_id = p_operation_id;
    IF FOUND THEN
        IF v_command IS DISTINCT FROM 'iam.invitation.expire' OR
           v_actor IS DISTINCT FROM p_actor_user_id OR
           v_request IS DISTINCT FROM p_request_sha256 OR
           v_id IS DISTINCT FROM p_invitation_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM invitation operation identity was reused';
        END IF;
        RETURN QUERY SELECT v_id, v_state, v_revision, true;
        RETURN;
    END IF;
    SELECT result.invitation_id, result.invitation_state,
           result.invitation_revision
    INTO v_id, v_state, v_revision
    FROM iam_invitation_expire_v1(
        p_scope_schema, p_invitation_id, p_expected_revision,
        p_actor_user_id, p_expected_authorization_version
    ) AS result;
    DELETE FROM iam_invitation_delivery_artifacts
    WHERE invitation_id = p_invitation_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM invitation delivery artifact is stale';
    END IF;
    INSERT INTO iam_invitation_operation_receipts (
        operation_id, command_key, actor_id, request_sha256, result_id,
        result_state, result_revision, created_at
    ) VALUES (
        p_operation_id, 'iam.invitation.expire', p_actor_user_id,
        p_request_sha256, v_id, v_state, v_revision,
        pg_catalog.transaction_timestamp()
    );
    RETURN QUERY SELECT v_id, v_state, v_revision, false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_invitation_expire_v2(
    name,uuid,bytea,uuid,bigint,uuid,bigint
) FROM PUBLIC;
