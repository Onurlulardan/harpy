-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_clock_read_v1()
RETURNS TABLE (database_now timestamptz)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing license clock read denied';
    END IF;
    RETURN QUERY SELECT pg_catalog.clock_timestamp();
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_clock_read_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_clock_read_v1()
TO harpy_runtime;
