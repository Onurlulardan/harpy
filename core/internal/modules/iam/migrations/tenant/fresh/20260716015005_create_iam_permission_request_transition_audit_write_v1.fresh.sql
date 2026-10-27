-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_transition_audit_write_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_request_sha256 bytea,
    p_group_id uuid,
    p_actor_user_id uuid,
    p_review_task_id uuid,
    p_command_key text,
    p_from_state text,
    p_to_state text,
    p_group_revision bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_group_state text;
    v_group_revision bigint;
    v_requester uuid;
    v_reviewer uuid;
    v_review_task uuid;
    v_manage_permission uuid;
    v_actor_active boolean;
    v_clock bigint;
    v_result_code text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request transition audit scope is invalid';
    END IF;
    IF p_operation_id IS NULL OR p_group_id IS NULL OR
       p_actor_user_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_group_revision < 1 OR p_command_key NOT IN (
           'iam.permission_request.create',
           'iam.permission_request.review',
           'iam.permission_request.cancel',
           'iam.permission_request.expire'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM permission request transition evidence is invalid';
    END IF;
    SELECT request_group.state, request_group.revision,
           request_group.requester_user_id, request_group.reviewer_user_id,
           request_group.review_task_id,
           request_group.creator_authority_permission_id
    INTO v_group_state, v_group_revision, v_requester, v_reviewer,
         v_review_task, v_manage_permission
    FROM iam_permission_request_groups AS request_group
    WHERE request_group.id = p_group_id;
    SELECT true INTO v_actor_active
    FROM iam_users AS actor
    WHERE actor.id = p_actor_user_id AND actor.status = 'active';
    IF v_group_state IS DISTINCT FROM p_to_state OR
       v_group_revision IS DISTINCT FROM p_group_revision OR
       v_actor_active IS DISTINCT FROM true OR
       (p_command_key = 'iam.permission_request.create' AND (
           p_from_state IS NOT NULL OR p_to_state <> 'PENDING' OR
           p_group_revision <> 1 OR p_review_task_id IS NOT NULL OR
           v_requester IS DISTINCT FROM p_actor_user_id)) OR
       (p_command_key = 'iam.permission_request.review' AND (
           p_from_state <> 'PENDING' OR p_to_state <> 'IN_REVIEW' OR
           p_review_task_id IS NULL OR
           v_reviewer IS DISTINCT FROM p_actor_user_id OR
           v_review_task IS DISTINCT FROM p_review_task_id)) OR
       (p_command_key = 'iam.permission_request.cancel' AND (
           p_from_state NOT IN ('PENDING', 'IN_REVIEW') OR
           p_to_state <> 'CANCELLED' OR p_review_task_id IS NOT NULL OR
           v_requester IS DISTINCT FROM p_actor_user_id)) OR
       (p_command_key = 'iam.permission_request.expire' AND (
           p_from_state NOT IN ('PENDING', 'IN_REVIEW') OR
           p_to_state <> 'EXPIRED' OR p_review_task_id IS NOT NULL OR
           NOT iam_user_has_permission_v1(
               p_actor_user_id, v_manage_permission
           ))) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM permission request transition evidence is stale';
    END IF;
    SELECT clock_value INTO v_clock
    FROM iam_authorization_clock
    WHERE singleton;
    IF v_clock IS NULL OR v_clock < 0 THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM authorization clock is unavailable';
    END IF;
    v_result_code := CASE p_command_key
        WHEN 'iam.permission_request.create'
            THEN 'iam.permission_request.created'
        WHEN 'iam.permission_request.review'
            THEN 'iam.permission_request.reviewed'
        WHEN 'iam.permission_request.cancel'
            THEN 'iam.permission_request.cancelled'
        ELSE 'iam.permission_request.expired'
    END;
    INSERT INTO iam_permission_request_transition_receipts (
        operation_id, command_key, request_sha256, result_code,
        result_id, result_revision, created_at
    ) VALUES (
        p_operation_id, p_command_key, p_request_sha256, v_result_code,
        p_group_id, p_group_revision,
        pg_catalog.transaction_timestamp()
    );
    INSERT INTO iam_permission_request_transition_audit (
        operation_id, group_id, actor_user_id, review_task_id, command_key,
        from_state, to_state, request_sha256, group_revision,
        authorization_clock, recorded_at
    ) VALUES (
        p_operation_id, p_group_id, p_actor_user_id, p_review_task_id,
        p_command_key, p_from_state, p_to_state, p_request_sha256,
        p_group_revision, v_clock, pg_catalog.transaction_timestamp()
    );
    RETURN v_clock;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_transition_audit_write_v1(
    name,uuid,bytea,uuid,uuid,uuid,text,text,text,bigint
) FROM PUBLIC;
