-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_reject_mutation()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $$
BEGIN
    RAISE EXCEPTION USING
        ERRCODE = '55000',
        MESSAGE = 'platform async history is immutable';
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_reject_mutation() FROM PUBLIC;
