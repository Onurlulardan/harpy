-- +goose Up
-- +goose StatementBegin
DROP FUNCTION platform.platform_async_redact_outbox_payloads_v1(integer);
CREATE FUNCTION platform.platform_async_redact_outbox_payloads_v1(
    p_batch_size integer
)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'unsafe outbox redaction v1 is retired; use v2';
END;
$function$;

DROP FUNCTION platform.platform_async_redact_dead_letter_payloads_v1(integer);
CREATE FUNCTION platform.platform_async_redact_dead_letter_payloads_v1(
    p_batch_size integer
)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'unsafe dead-letter redaction v1 is retired; use v2';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_redact_outbox_payloads_v1(integer)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.platform_async_redact_dead_letter_payloads_v1(integer)
FROM PUBLIC;
