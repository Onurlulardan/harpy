-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_resource_catalog_v1(
    p_scope_schema name,
    p_actor_user_id uuid,
    p_kind text,
    p_search text,
    p_after uuid,
    p_limit integer,
    p_missing_action_key text
)
RETURNS TABLE (
    resource_id uuid,
    resource_key text,
    display_name text,
    description text,
    catalog_kind text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request catalog scope is invalid';
    END IF;
    IF p_actor_user_id IS NULL OR NOT EXISTS (
        SELECT 1
        FROM iam_users AS actor
        WHERE actor.id OPERATOR(pg_catalog.=) p_actor_user_id
          AND actor.status OPERATOR(pg_catalog.=) 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request catalog actor is denied';
    END IF;
    IF p_kind IS NULL OR p_kind NOT IN ('all', 'form', 'system') OR
       p_search IS NULL OR p_search IS DISTINCT FROM pg_catalog.btrim(p_search) OR
       pg_catalog.char_length(p_search) > 128 OR p_limit IS NULL OR
       p_limit NOT BETWEEN 1 AND 101 OR
       (
           p_missing_action_key IS NOT NULL AND (
               pg_catalog.char_length(p_missing_action_key) NOT BETWEEN 1 AND 100 OR
               p_missing_action_key OPERATOR(pg_catalog.!~)
                   '^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$'
           )
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM permission request catalog query is invalid';
    END IF;
    IF p_missing_action_key IS NOT NULL AND NOT EXISTS (
        SELECT 1
        FROM iam_permissions AS permission
        JOIN iam_resources AS resource
          ON resource.id OPERATOR(pg_catalog.=) permission.resource_id
        JOIN iam_actions AS action
          ON action.id OPERATOR(pg_catalog.=) permission.action_id
        WHERE permission.match_kind OPERATOR(pg_catalog.=) 'exact'
          AND permission.state OPERATOR(pg_catalog.=) 'active'
          AND NOT permission.is_protected
          AND resource.state OPERATOR(pg_catalog.=) 'active'
          AND NOT resource.is_protected
          AND action.state OPERATOR(pg_catalog.=) 'active'
          AND NOT action.is_protected
          AND action.action_key OPERATOR(pg_catalog.=) p_missing_action_key
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM requestable action was not found';
    END IF;

    IF p_missing_action_key IS NULL THEN
        RETURN QUERY
        SELECT resource.id, resource.resource_key, resource.display_name,
               COALESCE(resource.description, ''),
               CASE WHEN pg_catalog.starts_with(resource.resource_key, 'form.')
                    THEN 'form'::text ELSE 'system'::text END
        FROM iam_resources AS resource
        WHERE resource.state OPERATOR(pg_catalog.=) 'active'
          AND NOT resource.is_protected
          AND (p_after IS NULL OR resource.id OPERATOR(pg_catalog.>) p_after)
          AND (
              p_kind OPERATOR(pg_catalog.=) 'all' OR
              (p_kind OPERATOR(pg_catalog.=) 'form' AND
               pg_catalog.starts_with(resource.resource_key, 'form.')) OR
              (p_kind OPERATOR(pg_catalog.=) 'system' AND
               NOT pg_catalog.starts_with(resource.resource_key, 'form.'))
          )
          AND (
              p_search OPERATOR(pg_catalog.=) '' OR
              pg_catalog.strpos(
                  pg_catalog.lower(pg_catalog.concat_ws(
                      ' ', resource.resource_key, resource.display_name,
                      COALESCE(resource.description, '')
                  )),
                  pg_catalog.lower(p_search)
              ) OPERATOR(pg_catalog.>) 0
          )
          AND EXISTS (
              SELECT 1
              FROM iam_permissions AS permission
              JOIN iam_actions AS action
                ON action.id OPERATOR(pg_catalog.=) permission.action_id
              WHERE permission.resource_id OPERATOR(pg_catalog.=) resource.id
                AND permission.match_kind OPERATOR(pg_catalog.=) 'exact'
                AND permission.state OPERATOR(pg_catalog.=) 'active'
                AND NOT permission.is_protected
                AND action.state OPERATOR(pg_catalog.=) 'active'
                AND NOT action.is_protected
          )
        ORDER BY resource.id
        LIMIT p_limit;
        RETURN;
    END IF;

    RETURN QUERY
    SELECT resource.id, resource.resource_key, resource.display_name,
           COALESCE(resource.description, ''),
           CASE WHEN pg_catalog.starts_with(resource.resource_key, 'form.')
                THEN 'form'::text ELSE 'system'::text END
    FROM iam_permissions AS permission
    JOIN iam_resources AS resource
      ON resource.id OPERATOR(pg_catalog.=) permission.resource_id
    JOIN iam_actions AS action
      ON action.id OPERATOR(pg_catalog.=) permission.action_id
    WHERE permission.match_kind OPERATOR(pg_catalog.=) 'exact'
      AND permission.state OPERATOR(pg_catalog.=) 'active'
      AND NOT permission.is_protected
      AND resource.state OPERATOR(pg_catalog.=) 'active'
      AND NOT resource.is_protected
      AND action.state OPERATOR(pg_catalog.=) 'active'
      AND NOT action.is_protected
      AND action.action_key OPERATOR(pg_catalog.=) p_missing_action_key
      AND (p_after IS NULL OR resource.id OPERATOR(pg_catalog.>) p_after)
      AND (
          p_kind OPERATOR(pg_catalog.=) 'all' OR
          (p_kind OPERATOR(pg_catalog.=) 'form' AND
           pg_catalog.starts_with(resource.resource_key, 'form.')) OR
          (p_kind OPERATOR(pg_catalog.=) 'system' AND
           NOT pg_catalog.starts_with(resource.resource_key, 'form.'))
      )
      AND (
          p_search OPERATOR(pg_catalog.=) '' OR
          pg_catalog.strpos(
              pg_catalog.lower(pg_catalog.concat_ws(
                  ' ', resource.resource_key, resource.display_name,
                  COALESCE(resource.description, '')
              )),
              pg_catalog.lower(p_search)
          ) OPERATOR(pg_catalog.>) 0
      )
      AND NOT iam_user_has_permission_v1(p_actor_user_id, permission.id)
    ORDER BY resource.id
    LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_resource_catalog_v1(
    name,uuid,text,text,uuid,integer,text
) FROM PUBLIC;
