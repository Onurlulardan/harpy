-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rest_query_actor_versions_v1(
    p_scope_schema name,
    p_actor_id uuid
)
RETURNS TABLE (
    token_version bigint,
    authorization_version bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_actor_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM REST Query actor scope is invalid';
    END IF;
    RETURN QUERY
    SELECT actor.token_version, actor.authorization_version
    FROM iam_users AS actor
    WHERE actor.id = p_actor_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rest_query_actor_versions_v1(name,uuid)
    FROM PUBLIC;
