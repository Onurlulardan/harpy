-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_initial_password_continuation_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_selector bytea
)
RETURNS SETOF iam_initial_password_continuations
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM initial password continuation scope is invalid';
    END IF;
    RETURN QUERY
    SELECT continuation.*
    FROM iam_initial_password_continuations AS continuation
    WHERE continuation.tenant_id OPERATOR(pg_catalog.=) p_tenant_id AND
          continuation.selector OPERATOR(pg_catalog.=) p_selector;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_initial_password_continuation_v1(
    name,uuid,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_initial_password_continuation_v1(
    name,uuid,bytea
) TO harpy_runtime;
