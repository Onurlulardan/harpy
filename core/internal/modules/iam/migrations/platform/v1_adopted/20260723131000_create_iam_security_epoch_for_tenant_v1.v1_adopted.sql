-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_security_epoch_for_tenant_v1(
    p_tenant_id uuid
)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_epoch bytea;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM tenant security epoch read is denied';
    END IF;
    SELECT epoch.security_epoch INTO STRICT v_epoch
    FROM platform.tenancy_tenant_bindings AS binding
    JOIN platform.iam_security_epochs AS epoch
      ON epoch.environment_id=binding.environment_id
    WHERE binding.tenant_id=p_tenant_id;
    IF pg_catalog.octet_length(v_epoch) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM tenant security epoch is unavailable';
    END IF;
    RETURN v_epoch;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_security_epoch_for_tenant_v1(uuid)
FROM PUBLIC;
