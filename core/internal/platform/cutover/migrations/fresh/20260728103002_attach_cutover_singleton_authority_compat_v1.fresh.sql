-- +goose Up
-- Migration stream version: 20260728103002.
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_sync_singleton_authority_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NEW.scope_kind <> 'tenant'
       OR NEW.domain NOT IN (
           'analytics', 'catalog', 'documents', 'work', 'workflow'
       ) THEN
        RETURN NEW;
    END IF;
    INSERT INTO platform.cutover_groups (
        scope_kind, scope_id, group_id, state, writer, generation,
        requested_by, approved_by, reason, change_reference,
        acceptance_digest, changed_at
    ) VALUES (
        NEW.scope_kind, NEW.scope_id, NEW.domain,
        NEW.state, NEW.writer, NEW.generation,
        NEW.requested_by, NEW.approved_by, NEW.reason,
        NEW.change_reference, NEW.acceptance_digest, NEW.changed_at
    )
    ON CONFLICT (scope_kind, scope_id, group_id) DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM platform.cutover_groups AS authority
        WHERE authority.scope_kind = NEW.scope_kind
          AND authority.scope_id = NEW.scope_id
          AND authority.group_id = NEW.domain
          AND authority.state IS NOT DISTINCT FROM NEW.state
          AND authority.writer IS NOT DISTINCT FROM NEW.writer
          AND authority.generation IS NOT DISTINCT FROM NEW.generation
          AND authority.requested_by IS NOT DISTINCT FROM NEW.requested_by
          AND authority.approved_by IS NOT DISTINCT FROM NEW.approved_by
          AND authority.reason IS NOT DISTINCT FROM NEW.reason
          AND authority.change_reference IS NOT DISTINCT FROM NEW.change_reference
          AND authority.acceptance_digest IS NOT DISTINCT FROM NEW.acceptance_digest
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'tenant singleton cutover group conflicts';
    END IF;
    INSERT INTO platform.cutover_group_members (
        scope_kind, scope_id, group_id, ordinal, domain
    ) VALUES (
        NEW.scope_kind, NEW.scope_id, NEW.domain, 1, NEW.domain
    )
    ON CONFLICT (scope_kind, scope_id, group_id, ordinal) DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM platform.cutover_group_members AS member
        WHERE member.scope_kind = NEW.scope_kind
          AND member.scope_id = NEW.scope_id
          AND member.group_id = NEW.domain
          AND member.ordinal = 1
          AND member.domain = NEW.domain
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'tenant singleton cutover member conflicts';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_sync_singleton_authority_v1()
    FROM PUBLIC;

CREATE TRIGGER cutover_sync_singleton_authority_insert_v1
AFTER INSERT ON platform.domain_cutovers
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_sync_singleton_authority_v1();
