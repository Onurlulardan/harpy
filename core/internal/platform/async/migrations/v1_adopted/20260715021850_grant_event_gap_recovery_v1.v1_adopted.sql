-- +goose Up
GRANT EXECUTE ON FUNCTION
    platform.platform_async_redact_outbox_payloads_v2(integer)
TO harpy_runtime;

GRANT EXECUTE ON FUNCTION
    platform.platform_async_redact_dead_letter_payloads_v2(integer)
TO harpy_runtime;
