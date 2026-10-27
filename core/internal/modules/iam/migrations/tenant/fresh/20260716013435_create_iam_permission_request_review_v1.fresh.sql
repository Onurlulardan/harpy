-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_review_v1(
    p_scope_schema name,
    p_group_id uuid,
    p_expected_revision bigint,
    p_expected_state text,
    p_reviewer_user_id uuid,
    p_expected_reviewer_authorization_version bigint,
    p_review_task_id uuid
)
RETURNS TABLE (group_id uuid, group_state text, group_revision bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_state text;
    v_revision bigint;
    v_expires_at timestamptz;
    v_reviewer_permission_id uuid;
    v_reviewer_state text;
    v_reviewer_authorization_version bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request review scope is invalid';
    END IF;
    IF p_group_id IS NULL OR p_expected_revision < 1 OR
       p_expected_state IS DISTINCT FROM 'PENDING' OR
       p_reviewer_user_id IS NULL OR p_review_task_id IS NULL OR
       p_expected_reviewer_authorization_version < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM permission request review input is invalid';
    END IF;
    SELECT request_group.state, request_group.revision,
           request_group.expires_at,
           request_group.reviewer_authority_permission_id,
           reviewer.status, reviewer.authorization_version
    INTO v_state, v_revision, v_expires_at, v_reviewer_permission_id,
         v_reviewer_state, v_reviewer_authorization_version
    FROM iam_permission_request_groups AS request_group
    JOIN iam_users AS reviewer ON reviewer.id = p_reviewer_user_id
    WHERE request_group.id OPERATOR(pg_catalog.=) p_group_id
    FOR UPDATE OF request_group, reviewer;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM permission request group is unknown';
    END IF;
    IF v_state IS DISTINCT FROM p_expected_state OR
       v_revision IS DISTINCT FROM p_expected_revision OR
       v_now >= v_expires_at OR v_reviewer_state IS DISTINCT FROM 'active' OR
       v_reviewer_authorization_version IS DISTINCT FROM
           p_expected_reviewer_authorization_version THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM permission request review CAS is stale';
    END IF;
    IF NOT iam_user_has_permission_v1(
        p_reviewer_user_id, v_reviewer_permission_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request reviewer authority is denied';
    END IF;
    UPDATE iam_permission_request_groups
    SET state = 'IN_REVIEW', reviewer_user_id = p_reviewer_user_id,
        review_task_id = p_review_task_id, revision = v_revision + 1,
        updated_at = v_now
    WHERE id = p_group_id;
    UPDATE iam_permission_requests AS request
    SET state = 'IN_REVIEW', revision = revision + 1, updated_at = v_now
    WHERE request.group_id = p_group_id;
    RETURN QUERY SELECT p_group_id, 'IN_REVIEW'::text, v_revision + 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_review_v1(
    name,uuid,bigint,text,uuid,bigint,uuid
) FROM PUBLIC;
