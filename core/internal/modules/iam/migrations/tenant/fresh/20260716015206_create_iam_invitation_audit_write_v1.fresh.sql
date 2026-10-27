-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_audit_write_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_invitation_id uuid, p_request_sha256 bytea, p_command_key text,
    p_result_state text, p_result_revision bigint,
    p_delivery_artifact_ref text, p_authorization_clock bigint
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_command text;
    v_actor uuid;
    v_request bytea;
    v_result_id uuid;
    v_state text;
    v_revision bigint;
    v_clock bigint;
    v_result_code text;
    v_invitation_state text;
    v_invitation_revision bigint;
    v_actor_active boolean;
    v_artifact_ref text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_operation_id IS NULL OR p_actor_id IS NULL OR
       p_invitation_id IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM invitation audit scope is invalid';
    END IF;
    SELECT state, revision INTO v_invitation_state, v_invitation_revision
    FROM iam_invitations WHERE id = p_invitation_id;
    SELECT true INTO v_actor_active FROM iam_users
    WHERE id = p_actor_id AND status = 'active';
    IF p_command_key = 'iam.invitation.accept' THEN
        SELECT command_key, request_sha256, result_id, result_code,
               result_revision, clock_value
        INTO v_command, v_request, v_result_id, v_result_code,
             v_revision, v_clock
        FROM iam_authorization_operation_receipts
        WHERE operation_id = p_operation_id;
        v_actor := p_actor_id;
        v_state := 'accepted';
        IF v_result_code IS DISTINCT FROM 'iam.invitation.accepted' OR
           v_clock IS DISTINCT FROM p_authorization_clock THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'IAM invitation acceptance audit evidence is invalid';
        END IF;
    ELSE
        SELECT command_key, actor_id, request_sha256, result_id,
               result_state, result_revision
        INTO v_command, v_actor, v_request, v_result_id, v_state, v_revision
        FROM iam_invitation_operation_receipts
        WHERE operation_id = p_operation_id;
        IF p_authorization_clock IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'IAM invitation audit clock is invalid';
        END IF;
    END IF;
    IF p_command_key = 'iam.invitation.create' THEN
        SELECT artifact_ref INTO v_artifact_ref
        FROM iam_invitation_delivery_artifacts
        WHERE invitation_id = p_invitation_id;
    END IF;
    IF v_command IS DISTINCT FROM p_command_key OR
       v_actor IS DISTINCT FROM p_actor_id OR
       v_request IS DISTINCT FROM p_request_sha256 OR
       v_result_id IS DISTINCT FROM p_invitation_id OR
       v_state IS DISTINCT FROM p_result_state OR
       v_revision IS DISTINCT FROM p_result_revision OR
       v_invitation_state IS DISTINCT FROM p_result_state OR
       v_invitation_revision IS DISTINCT FROM p_result_revision OR
       v_actor_active IS DISTINCT FROM true OR
       v_artifact_ref IS DISTINCT FROM p_delivery_artifact_ref THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM invitation audit evidence is invalid';
    END IF;
    INSERT INTO iam_invitation_change_audit (
        operation_id, command_key, actor_id, invitation_id, request_sha256,
        result_state, result_revision, delivery_artifact_ref,
        authorization_clock, recorded_at
    ) VALUES (
        p_operation_id, p_command_key, p_actor_id, p_invitation_id,
        p_request_sha256, p_result_state, p_result_revision,
        p_delivery_artifact_ref, p_authorization_clock,
        pg_catalog.transaction_timestamp()
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_invitation_audit_write_v1(
    name,uuid,uuid,uuid,bytea,text,text,bigint,text,bigint
) FROM PUBLIC;

