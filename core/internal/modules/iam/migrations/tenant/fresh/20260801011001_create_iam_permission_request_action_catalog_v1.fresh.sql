-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_action_catalog_v1(
    p_scope_schema name,
    p_actor_user_id uuid,
    p_search text,
    p_after uuid,
    p_limit integer
)
RETURNS TABLE (
    action_id uuid,
    action_key text,
    display_name text,
    description text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request action catalog scope is invalid';
    END IF;
    IF p_actor_user_id IS NULL OR NOT EXISTS (
        SELECT 1
        FROM iam_users AS actor
        WHERE actor.id OPERATOR(pg_catalog.=) p_actor_user_id
          AND actor.status OPERATOR(pg_catalog.=) 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request action catalog actor is denied';
    END IF;
    IF p_search IS NULL OR p_search IS DISTINCT FROM pg_catalog.btrim(p_search) OR
       pg_catalog.char_length(p_search) > 128 OR p_limit IS NULL OR
       p_limit NOT BETWEEN 1 AND 101 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM permission request action catalog query is invalid';
    END IF;

    RETURN QUERY
    SELECT action.id, action.action_key, action.display_name,
           COALESCE(action.description, '')
    FROM iam_actions AS action
    WHERE action.state OPERATOR(pg_catalog.=) 'active'
      AND NOT action.is_protected
      AND (p_after IS NULL OR action.id OPERATOR(pg_catalog.>) p_after)
      AND (
          p_search OPERATOR(pg_catalog.=) '' OR
          pg_catalog.strpos(
              pg_catalog.lower(pg_catalog.concat_ws(
                  ' ', action.action_key, action.display_name,
                  COALESCE(action.description, '')
              )),
              pg_catalog.lower(p_search)
          ) OPERATOR(pg_catalog.>) 0
      )
      AND EXISTS (
          SELECT 1
          FROM iam_permissions AS permission
          JOIN iam_resources AS resource
            ON resource.id OPERATOR(pg_catalog.=) permission.resource_id
          WHERE permission.action_id OPERATOR(pg_catalog.=) action.id
            AND permission.match_kind OPERATOR(pg_catalog.=) 'exact'
            AND permission.state OPERATOR(pg_catalog.=) 'active'
            AND NOT permission.is_protected
            AND resource.state OPERATOR(pg_catalog.=) 'active'
            AND NOT resource.is_protected
      )
    ORDER BY action.id
    LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_action_catalog_v1(
    name,uuid,text,uuid,integer
) FROM PUBLIC;
