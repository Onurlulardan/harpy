-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_owner_insert_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       NOT (TG_TABLE_NAME OPERATOR(pg_catalog.=) ANY (
           ARRAY['jobs','outbox','idempotency_receipts','dead_letters']::text[]
       )) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5604',
            MESSAGE = 'async owner insert trigger target is invalid';
    END IF;

    IF NOT (
        (SESSION_USER OPERATOR(pg_catalog.=) 'forgecrud_v2_runtime' AND
            NEW.writer_origin OPERATOR(pg_catalog.=) 'v2') OR
        (SESSION_USER OPERATOR(pg_catalog.=) 'forgecrud_adopter' AND
            NEW.writer_origin OPERATOR(pg_catalog.=) 'v1')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5605',
            MESSAGE = 'async owner insert authority is invalid';
    END IF;

    IF platform.platform_async_assert_current_authority_v1(
        NEW.scope_kind,
        NEW.scope_id,
        NEW.authority_group_id,
        NEW.authority_group_generation,
        NEW.authority_domain,
        NEW.cutover_generation
    ) IS NULL THEN
        NULL;
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_validate_owner_insert_v1()
FROM PUBLIC;

CREATE TRIGGER jobs_authority_insert_guard
BEFORE INSERT ON platform.jobs
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v1();
ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_authority_insert_guard;

CREATE TRIGGER outbox_authority_insert_guard
BEFORE INSERT ON platform.outbox
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v1();
ALTER TABLE platform.outbox ENABLE ALWAYS TRIGGER outbox_authority_insert_guard;

CREATE TRIGGER idempotency_receipts_authority_insert_guard
BEFORE INSERT ON platform.idempotency_receipts
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v1();
ALTER TABLE platform.idempotency_receipts
    ENABLE ALWAYS TRIGGER idempotency_receipts_authority_insert_guard;

CREATE TRIGGER dead_letters_authority_insert_guard
BEFORE INSERT ON platform.dead_letters
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v1();
ALTER TABLE platform.dead_letters
    ENABLE ALWAYS TRIGGER dead_letters_authority_insert_guard;
