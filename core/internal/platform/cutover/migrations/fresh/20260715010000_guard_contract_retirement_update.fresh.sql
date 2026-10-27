-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.contract_retirement_guard_update()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF OLD.state OPERATOR(pg_catalog.=) 'retired' OR
       NEW.contract_key IS DISTINCT FROM OLD.contract_key OR
       NEW.owner_module IS DISTINCT FROM OLD.owner_module OR
       NEW.domain IS DISTINCT FROM OLD.domain OR
       NEW.revision IS DISTINCT FROM OLD.revision OPERATOR(pg_catalog.+) 1 OR
       NEW.updated_at OPERATOR(pg_catalog.<) OLD.updated_at OR
       NEW.updated_at OPERATOR(pg_catalog.>) pg_catalog.clock_timestamp() OR
       NEW.hit_count OPERATOR(pg_catalog.<) OLD.hit_count OR
       (OLD.last_hit_at IS NOT NULL AND
        NEW.last_hit_at OPERATOR(pg_catalog.<) OLD.last_hit_at) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'contract retirement monotonic update rejected';
    END IF;
    IF NEW.hit_count IS DISTINCT FROM OLD.hit_count THEN
        IF NEW.hit_count IS DISTINCT FROM OLD.hit_count OPERATOR(pg_catalog.+) 1 OR
           NEW.last_hit_at IS DISTINCT FROM NEW.updated_at OR
           NEW.state IS DISTINCT FROM 'active' OR
           NEW.remaining_scope_count IS DISTINCT FROM OLD.remaining_scope_count OR
           NEW.zero_hit_since IS NOT NULL OR NEW.eligible_at IS NOT NULL OR
           NEW.retired_at IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'contract retirement hit update rejected';
        END IF;
    ELSIF NEW.last_hit_at IS DISTINCT FROM OLD.last_hit_at THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'contract retirement last hit update rejected';
    END IF;
    IF NEW.state OPERATOR(pg_catalog.=) 'active' THEN
        IF NEW.eligible_at IS NOT NULL OR NEW.retired_at IS NOT NULL OR
           (NEW.remaining_scope_count OPERATOR(pg_catalog.>) 0 AND
            NEW.zero_hit_since IS NOT NULL) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'contract retirement active state rejected';
        END IF;
    ELSIF NEW.state OPERATOR(pg_catalog.=) 'eligible' THEN
        IF NEW.remaining_scope_count IS DISTINCT FROM 0 OR
           NEW.zero_hit_since IS NULL OR NEW.eligible_at IS NULL OR
           NEW.retired_at IS NOT NULL OR
           NEW.eligible_at OPERATOR(pg_catalog.<) (
               NEW.zero_hit_since OPERATOR(pg_catalog.+) INTERVAL '30 days'
           ) OR
           (OLD.state IS DISTINCT FROM 'eligible' AND
            NEW.eligible_at IS DISTINCT FROM NEW.updated_at) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'contract retirement eligible state rejected';
        END IF;
    ELSIF NEW.state OPERATOR(pg_catalog.=) 'retired' THEN
        IF OLD.state IS DISTINCT FROM 'eligible' OR
           NEW.remaining_scope_count IS DISTINCT FROM 0 OR
           NEW.zero_hit_since IS NULL OR NEW.eligible_at IS NULL OR
           NEW.retired_at IS NULL OR
           NEW.retired_at IS DISTINCT FROM NEW.updated_at OR
           NEW.retired_at OPERATOR(pg_catalog.<) NEW.eligible_at OR
           NEW.retired_at OPERATOR(pg_catalog.<) (
               NEW.zero_hit_since OPERATOR(pg_catalog.+) INTERVAL '30 days'
           ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'contract retirement terminal state rejected';
        END IF;
    ELSE
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'contract retirement state rejected';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER contract_retirements_update_guard
BEFORE UPDATE ON platform.contract_retirements
FOR EACH ROW
EXECUTE FUNCTION platform.contract_retirement_guard_update();
