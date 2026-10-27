-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_work_available_users_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_search text,
    p_limit integer
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_evidence jsonb;
BEGIN
    IF p_search IS NULL OR pg_catalog.char_length(p_search) > 200 OR
       p_limit NOT BETWEEN 1 AND 50 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM Work directory input is invalid';
    END IF;
    SELECT actor_principals INTO v_evidence
    FROM iam_work_principal_evidence_v1(
        p_scope_schema, p_tenant_id, p_actor_id, 'USER', p_actor_id
    );
    IF v_evidence IS NULL OR NOT EXISTS (
        SELECT 1
        FROM iam_users AS actor
        JOIN iam_user_access_modes AS access_mode
          ON access_mode.user_id = actor.id
         AND access_mode.mode = 'operator_console'
         AND access_mode.state = 'active'
        WHERE actor.id = p_actor_id AND actor.status = 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM Work directory actor is invalid';
    END IF;
    RETURN COALESCE((
        WITH users AS (
            SELECT subject.id,
                COALESCE(identifier.display_value, '') AS email,
                COALESCE(profile.first_name, profile.display_name, '') AS first_name,
                COALESCE(profile.last_name, '') AS last_name,
                COALESCE(profile.avatar, '') AS avatar
            FROM iam_users AS subject
            LEFT JOIN LATERAL (
                SELECT value.display_value
                FROM iam_user_identifiers AS value
                WHERE value.user_id = subject.id AND value.kind = 'email_v1'
                  AND value.state = 'active'
                ORDER BY value.updated_at DESC, value.id LIMIT 1
            ) AS identifier ON true
            LEFT JOIN LATERAL (
                SELECT
                    pg_catalog.max(subject.profile_values ->> field.field_id::text)
                        FILTER (WHERE field.field_key = 'first_name') AS first_name,
                    pg_catalog.max(subject.profile_values ->> field.field_id::text)
                        FILTER (WHERE field.field_key = 'last_name') AS last_name,
                    pg_catalog.max(subject.profile_values ->> field.field_id::text)
                        FILTER (WHERE field.field_key = 'display_name') AS display_name,
                    pg_catalog.max(subject.profile_values ->> field.field_id::text)
                        FILTER (WHERE field.field_key = 'avatar') AS avatar
                FROM iam_profile_field_revisions AS field
                WHERE field.schema_version_id = subject.profile_schema_version_id
                  AND field.resource_kind = 'user' AND field.field_state = 'active'
                  AND field.field_key IN ('first_name','last_name','display_name','avatar')
            ) AS profile ON true
            WHERE subject.status = 'active'
        ), filtered AS (
            SELECT users.* FROM users
            WHERE p_search = '' OR pg_catalog.lower(users.email) LIKE
                    '%' || pg_catalog.lower(p_search) || '%' OR
                pg_catalog.lower(users.first_name) LIKE
                    '%' || pg_catalog.lower(p_search) || '%' OR
                pg_catalog.lower(users.last_name) LIKE
                    '%' || pg_catalog.lower(p_search) || '%'
            ORDER BY users.first_name, users.last_name, users.email, users.id
            LIMIT p_limit
        )
        SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
            'id', filtered.id, 'email', filtered.email,
            'first_name', filtered.first_name, 'last_name', filtered.last_name,
            'avatar', filtered.avatar
        ) ORDER BY filtered.first_name, filtered.last_name, filtered.email, filtered.id)
        FROM filtered
    ), '[]'::jsonb);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_work_available_users_v1(name,uuid,uuid,text,integer)
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_work_available_users_v1(name,uuid,uuid,text,integer)
    TO forgecrud_runtime;
