-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_hmac_key_id_valid_v1(p_value text)
RETURNS boolean
LANGUAGE sql
IMMUTABLE
PARALLEL SAFE
SET search_path TO pg_catalog
AS $function$
SELECT p_value IS NOT NULL AND p_value OPERATOR(pg_catalog.~)
    '^(([A-Za-z0-9_-]{4}){4,16}|([A-Za-z0-9_-]{4}){4,15}[A-Za-z0-9_-][AQgw]|([A-Za-z0-9_-]{4}){4,15}[A-Za-z0-9_-]{2}[AEIMQUYcgkosw048])$'
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_hmac_key_id_valid_v1(text) FROM PUBLIC;
