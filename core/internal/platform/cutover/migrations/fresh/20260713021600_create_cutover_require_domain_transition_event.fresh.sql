-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_require_domain_transition_event()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM platform.cutover_transition_events AS event
        WHERE event.event_xid = pg_catalog.pg_current_xact_id()
          AND event.target_kind = 'domain'
          AND event.scope_kind = NEW.scope_kind
          AND event.scope_id = NEW.scope_id
          AND event.domain = NEW.domain
          AND event.previous_state = OLD.state
          AND event.new_state = NEW.state
          AND event.previous_writer = OLD.writer
          AND event.new_writer = NEW.writer
          AND event.previous_generation = OLD.generation
          AND event.new_generation = NEW.generation
          AND event.requested_by = NEW.requested_by
          AND event.approved_by IS NOT DISTINCT FROM NEW.approved_by
          AND event.reason = NEW.reason
          AND event.change_reference = NEW.change_reference
          AND event.acceptance_digest IS NOT DISTINCT FROM NEW.acceptance_digest
          AND event.occurred_at = NEW.changed_at
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'domain cutover update requires matching immutable transition evidence';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM platform.cutover_group_members AS member
        JOIN platform.cutover_groups AS cutover_group
          ON cutover_group.scope_kind = member.scope_kind
         AND cutover_group.scope_id = member.scope_id
         AND cutover_group.group_id = member.group_id
        WHERE member.scope_kind = NEW.scope_kind
          AND member.scope_id = NEW.scope_id
          AND member.domain = NEW.domain
          AND cutover_group.state = NEW.state
          AND cutover_group.writer = NEW.writer
          AND cutover_group.generation = NEW.generation
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'domain cutover update must remain synchronized with its cutover group';
    END IF;

    RETURN NULL;
END;
$function$;
-- +goose StatementEnd
