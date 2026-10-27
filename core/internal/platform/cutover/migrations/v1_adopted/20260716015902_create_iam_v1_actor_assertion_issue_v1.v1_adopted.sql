-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_v1_actor_assertion_issue_v1(
    p_tenant_id uuid,
    p_target_domain text
)
RETURNS TABLE (
    iam_group_generation bigint,
    iam_generation bigint,
    target_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 actor assertion issue authority is denied';
    END IF;
    RETURN QUERY
    SELECT authority.iam_group_generation, authority.iam_generation,
           authority.target_generation
    FROM platform.iam_v1_actor_bridge_authority_v1(
        p_tenant_id, p_target_domain
    ) AS authority;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_v1_actor_assertion_issue_v1(uuid,text)
FROM PUBLIC;
GRANT USAGE ON SCHEMA platform TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_v1_actor_assertion_issue_v1(uuid,text)
TO harpy_runtime;

