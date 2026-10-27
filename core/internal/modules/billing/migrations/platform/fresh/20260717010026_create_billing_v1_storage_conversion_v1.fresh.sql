-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_convert_v1_storage_megabytes_v1(
    p_megabytes bigint
)
RETURNS TABLE (is_unlimited boolean, storage_bytes bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF p_megabytes IS NULL OR p_megabytes < -1
       OR p_megabytes > 8583068847 THEN
        RAISE EXCEPTION USING ERRCODE = '22003',
            MESSAGE = 'billing V1 storage limit is outside canonical bounds';
    END IF;
    IF p_megabytes IS NOT DISTINCT FROM -1 THEN
        RETURN QUERY SELECT true, NULL::bigint;
        RETURN;
    END IF;
    RETURN QUERY SELECT false, p_megabytes * 1048576;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_convert_v1_storage_megabytes_v1(bigint)
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_convert_v1_storage_megabytes_v1(bigint)
    TO harpy_adopter;
-- +goose StatementEnd
