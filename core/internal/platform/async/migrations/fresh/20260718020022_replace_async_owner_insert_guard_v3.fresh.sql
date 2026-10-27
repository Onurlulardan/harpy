-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_owner_insert_v3()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT'
       OR TG_TABLE_SCHEMA IS DISTINCT FROM 'platform'
       OR NOT (TG_TABLE_NAME OPERATOR(pg_catalog.=) ANY (
            ARRAY['jobs','outbox','idempotency_receipts','dead_letters']::text[]
       )) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5604',
            MESSAGE = 'async owner insert trigger target is invalid';
    END IF;

    IF TG_TABLE_NAME OPERATOR(pg_catalog.=) 'jobs'
       AND SESSION_USER OPERATOR(pg_catalog.=) 'forgecrud_adopter'
       AND NEW.writer_origin OPERATOR(pg_catalog.=) 'v1' THEN
        PERFORM platform.platform_async_assert_v1_adopted_job_v1(NEW);
        RETURN NEW;
    END IF;

    IF TG_TABLE_NAME OPERATOR(pg_catalog.=) 'jobs'
       AND SESSION_USER OPERATOR(pg_catalog.=) 'forgecrud_migrator'
       AND NEW.writer_origin OPERATOR(pg_catalog.=) 'v2' THEN
        PERFORM
            platform.platform_async_assert_forms_schema_completion_job_v1(NEW);
        RETURN NEW;
    END IF;

    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR NEW.writer_origin IS DISTINCT FROM 'v2' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5605',
            MESSAGE = 'async owner insert authority is invalid';
    END IF;
    PERFORM platform.platform_async_assert_current_authority_v1(
        NEW.scope_kind, NEW.scope_id, NEW.authority_group_id,
        NEW.authority_group_generation, NEW.authority_domain,
        NEW.cutover_generation
    );
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_validate_owner_insert_v3()
FROM PUBLIC;

CREATE OR REPLACE TRIGGER jobs_authority_insert_guard
BEFORE INSERT ON platform.jobs FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v3();
ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_authority_insert_guard;

CREATE OR REPLACE TRIGGER outbox_authority_insert_guard
BEFORE INSERT ON platform.outbox FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v3();
ALTER TABLE platform.outbox ENABLE ALWAYS TRIGGER outbox_authority_insert_guard;

CREATE OR REPLACE TRIGGER idempotency_receipts_authority_insert_guard
BEFORE INSERT ON platform.idempotency_receipts FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v3();
ALTER TABLE platform.idempotency_receipts
    ENABLE ALWAYS TRIGGER idempotency_receipts_authority_insert_guard;

CREATE OR REPLACE TRIGGER dead_letters_authority_insert_guard
BEFORE INSERT ON platform.dead_letters FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v3();
ALTER TABLE platform.dead_letters
    ENABLE ALWAYS TRIGGER dead_letters_authority_insert_guard;

