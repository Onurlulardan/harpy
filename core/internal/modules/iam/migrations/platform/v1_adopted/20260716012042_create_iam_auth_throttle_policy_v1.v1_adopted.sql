-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_auth_throttle_policy_v1(
    p_scope_kind text,
    p_purpose text,
    p_dimension text,
    p_policy_version integer,
    p_window_seconds integer,
    p_block_after integer,
    p_block_seconds integer
)
RETURNS boolean
LANGUAGE sql
IMMUTABLE
PARALLEL SAFE
SET search_path = pg_catalog
AS $function$
SELECT
    p_policy_version IS NOT DISTINCT FROM 1 AND
    p_window_seconds IS NOT DISTINCT FROM CASE
        WHEN p_purpose OPERATOR(pg_catalog.=) ANY (
            ARRAY['login','refresh','reauth']::text[]
        ) THEN 300
        WHEN p_purpose OPERATOR(pg_catalog.=) ANY (
            ARRAY['password_reset','email_verification',
                  'portal_setup','invitation']::text[]
        ) THEN 900
    END AND
    p_block_seconds IS NOT DISTINCT FROM CASE
        WHEN p_purpose OPERATOR(pg_catalog.=) ANY (
            ARRAY['login','refresh','reauth']::text[]
        ) THEN 900
        WHEN p_purpose OPERATOR(pg_catalog.=) ANY (
            ARRAY['password_reset','email_verification',
                  'portal_setup','invitation']::text[]
        ) THEN 3600
    END AND
    p_block_after IS NOT DISTINCT FROM CASE
        WHEN p_purpose OPERATOR(pg_catalog.=) ANY (
                 ARRAY['login','refresh','reauth']::text[]
             ) AND p_dimension OPERATOR(pg_catalog.=) ANY (
                 ARRAY['identifier','selector']::text[]
             ) THEN 5
        WHEN p_purpose OPERATOR(pg_catalog.=) ANY (
                 ARRAY['password_reset','email_verification',
                       'portal_setup','invitation']::text[]
             ) AND p_dimension OPERATOR(pg_catalog.=) 'identifier' THEN 3
        WHEN p_dimension OPERATOR(pg_catalog.=) 'identifier_ip' THEN 10
        WHEN p_dimension OPERATOR(pg_catalog.=) 'tenant_slug' THEN 30
        WHEN p_dimension OPERATOR(pg_catalog.=) 'ip' THEN 100
    END AND
    (
        p_scope_kind OPERATOR(pg_catalog.=) 'tenant' AND
        (
            p_purpose OPERATOR(pg_catalog.=) 'login' AND
            p_dimension OPERATOR(pg_catalog.=) ANY (
                ARRAY['identifier','ip','identifier_ip']::text[]
            ) OR
            p_purpose OPERATOR(pg_catalog.=) 'refresh' AND
            p_dimension OPERATOR(pg_catalog.=) ANY (
                ARRAY['selector','ip']::text[]
            ) OR
            p_purpose OPERATOR(pg_catalog.=) 'reauth' AND
            p_dimension OPERATOR(pg_catalog.=) ANY (
                ARRAY['identifier','ip','identifier_ip']::text[]
            ) OR
            p_purpose OPERATOR(pg_catalog.=) ANY (
                ARRAY['password_reset','email_verification',
                      'portal_setup','invitation']::text[]
            ) AND p_dimension OPERATOR(pg_catalog.=) ANY (
                ARRAY['identifier','ip','identifier_ip']::text[]
            )
        ) OR
        p_scope_kind OPERATOR(pg_catalog.=) 'environment' AND
        (
            p_purpose OPERATOR(pg_catalog.=) 'login' AND
            p_dimension OPERATOR(pg_catalog.=) ANY (
                ARRAY['tenant_slug','identifier','ip','identifier_ip']::text[]
            ) OR
            p_purpose OPERATOR(pg_catalog.=) 'refresh' AND
            p_dimension OPERATOR(pg_catalog.=) ANY (
                ARRAY['selector','ip']::text[]
            ) OR
            p_purpose OPERATOR(pg_catalog.=) ANY (
                ARRAY['password_reset','email_verification',
                      'portal_setup','invitation']::text[]
            ) AND p_dimension OPERATOR(pg_catalog.=) ANY (
                ARRAY['tenant_slug','identifier','ip','identifier_ip']::text[]
            )
        )
    );
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_auth_throttle_policy_v1(
    text,text,text,integer,integer,integer,integer
) FROM PUBLIC;
