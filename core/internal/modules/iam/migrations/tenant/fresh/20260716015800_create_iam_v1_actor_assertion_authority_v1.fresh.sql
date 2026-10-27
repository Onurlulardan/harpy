-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_actor_assertion_authority_v1(
    p_scope_schema name,
    p_subject_id uuid,
    p_permission_id uuid,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint,
    p_expected_authorization_clock bigint
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_clock bigint;
    v_policy_epoch bigint;
    v_status text;
    v_token_version bigint;
    v_authorization_version bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v1_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 actor assertion scope is invalid';
    END IF;
    IF p_subject_id IS NULL OR p_permission_id IS NULL OR
       p_expected_token_version < 1 OR
       p_expected_authorization_version < 1 OR
       p_expected_authorization_clock < 0 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 actor assertion evidence is invalid';
    END IF;

    SELECT clock_value INTO v_clock
    FROM iam_authorization_clock
    WHERE singleton
    FOR SHARE;
    IF NOT FOUND OR v_clock IS DISTINCT FROM p_expected_authorization_clock THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 actor assertion clock is stale';
    END IF;

    SELECT authorization_epoch INTO v_policy_epoch
    FROM iam_protected_admin_policies
    WHERE singleton
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 actor assertion policy is unavailable';
    END IF;

    SELECT status, token_version, authorization_version
    INTO v_status, v_token_version, v_authorization_version
    FROM iam_users
    WHERE id = p_subject_id
    FOR UPDATE;
    IF NOT FOUND OR v_status IS DISTINCT FROM 'active' OR
       v_token_version IS DISTINCT FROM p_expected_token_version OR
       GREATEST(v_authorization_version, v_policy_epoch)
           IS DISTINCT FROM p_expected_authorization_version OR
       NOT iam_user_has_permission_v1(p_subject_id, p_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 actor assertion authority is denied';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_actor_assertion_authority_v1(
    name,uuid,uuid,bigint,bigint,bigint
) FROM PUBLIC;
