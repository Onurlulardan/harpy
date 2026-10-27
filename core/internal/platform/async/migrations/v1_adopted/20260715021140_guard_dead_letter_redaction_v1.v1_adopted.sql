-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_guard_dead_letter_redaction_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
DECLARE
    expected platform.dead_letters%ROWTYPE;
    redaction_time timestamptz;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'dead_letters' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5649',
            MESSAGE = 'dead-letter redaction trigger target is invalid';
    END IF;

    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       OLD.record_kind IS DISTINCT FROM 'root' OR
       NEW.record_kind IS DISTINCT FROM 'root' OR
       OLD.payload_canonical IS NULL OR
       OLD.payload_redacted_at IS NOT NULL OR
       NEW.payload_canonical IS NOT NULL OR
       NEW.payload_redacted_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'dead-letter record is immutable';
    END IF;

    redaction_time := pg_catalog.clock_timestamp();
    IF redaction_time < OLD.full_payload_expires_at THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'dead-letter payload retention has not expired';
    END IF;

    NEW.payload_redacted_at := redaction_time;
    expected := OLD;
    expected.payload_canonical := NULL;
    expected.payload_redacted_at := redaction_time;
    IF NEW IS DISTINCT FROM expected THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'dead-letter redaction changed immutable evidence';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_dead_letter_redaction_v1()
FROM PUBLIC;

CREATE TRIGGER dead_letters_redaction_guard
BEFORE UPDATE ON platform.dead_letters
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_dead_letter_redaction_v1();
ALTER TABLE platform.dead_letters
    ENABLE ALWAYS TRIGGER dead_letters_redaction_guard;
