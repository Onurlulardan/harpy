-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_outbox_authority_current_v1(
    p_event platform.outbox
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_group_state text;
    v_group_writer text;
    v_group_generation bigint;
    v_domain_state text;
    v_domain_writer text;
    v_domain_generation bigint;
    v_member boolean;
BEGIN
    SELECT authority_group.state,authority_group.writer,
        authority_group.generation
    INTO v_group_state,v_group_writer,v_group_generation
    FROM platform.cutover_groups AS authority_group
    WHERE authority_group.scope_kind = p_event.scope_kind
      AND authority_group.scope_id = p_event.scope_id
      AND authority_group.group_id = p_event.authority_group_id
    FOR SHARE;
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    SELECT authority_domain.state,authority_domain.writer,
        authority_domain.generation
    INTO v_domain_state,v_domain_writer,v_domain_generation
    FROM platform.domain_cutovers AS authority_domain
    WHERE authority_domain.scope_kind = p_event.scope_kind
      AND authority_domain.scope_id = p_event.scope_id
      AND authority_domain.domain = p_event.authority_domain
    FOR SHARE;
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    SELECT true INTO v_member FROM platform.cutover_group_members AS member
    WHERE member.scope_kind = p_event.scope_kind
      AND member.scope_id = p_event.scope_id
      AND member.group_id = p_event.authority_group_id
      AND member.domain = p_event.authority_domain
    FOR SHARE;
    RETURN p_event.parked_at IS NULL AND p_event.writer_origin = 'v2'
        AND v_group_state = 'v2' AND v_group_writer = 'v2'
        AND v_group_generation = p_event.authority_group_generation
        AND v_domain_state = 'v2' AND v_domain_writer = 'v2'
        AND v_domain_generation = p_event.cutover_generation
        AND COALESCE(v_member,false);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_outbox_authority_current_v1(platform.outbox)
FROM PUBLIC;
