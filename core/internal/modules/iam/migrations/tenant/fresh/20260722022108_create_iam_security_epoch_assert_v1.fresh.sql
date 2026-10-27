-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_security_epoch_assert_v1(
    p_scope_schema name, p_expected_security_epoch bytea
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.octet_length(p_expected_security_epoch) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM security epoch assertion input is invalid';
    END IF;
    IF p_expected_security_epoch IS DISTINCT FROM
       iam_security_epoch_current_v1(p_scope_schema) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM security epoch evidence is stale';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_security_epoch_assert_v1(name,bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_security_epoch_assert_v1(name,bytea)
TO forgecrud_v2_runtime;
