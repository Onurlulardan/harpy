-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_security_epoch_current_v1(p_scope_schema name)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_epoch bytea;
BEGIN
    IF p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM security epoch scope is invalid';
    END IF;
    SELECT platform.iam_security_epoch_for_tenant_v1(identity.tenant_id)
    INTO v_epoch
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton;
    IF pg_catalog.octet_length(v_epoch) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM security epoch is unavailable';
    END IF;
    RETURN v_epoch;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_security_epoch_current_v1(name) FROM PUBLIC;
