-- +goose Up
-- Migration stream version: 20260802235005.
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_ensure_tenancy_authority_v1(p_tenant_id uuid)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF p_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'tenancy cutover tenant is required';
    END IF;
    INSERT INTO platform.domain_cutovers (
        scope_kind, scope_id, domain, state, writer, generation,
        requested_by, approved_by, reason, change_reference,
        acceptance_digest, changed_at
    )
    SELECT domain.scope_kind, domain.scope_id, seed.domain,
           domain.state, domain.writer, domain.generation,
           domain.requested_by, domain.approved_by, domain.reason,
           domain.change_reference, domain.acceptance_digest,
           domain.changed_at
    FROM platform.domain_cutovers AS domain
    CROSS JOIN pg_catalog.unnest(ARRAY['tenancy']) AS seed(domain)
    WHERE domain.scope_kind = 'tenant'
      AND domain.scope_id = p_tenant_id
      AND domain.domain = 'workflow'
      AND NOT EXISTS (
          SELECT 1 FROM platform.domain_cutovers AS existing
          WHERE existing.scope_kind = domain.scope_kind
            AND existing.scope_id = domain.scope_id
            AND existing.domain = seed.domain
      );
    INSERT INTO platform.cutover_groups (
        scope_kind, scope_id, group_id, state, writer, generation,
        requested_by, approved_by, reason, change_reference,
        acceptance_digest, changed_at
    )
    SELECT domain.scope_kind, domain.scope_id, domain.domain,
           domain.state, domain.writer, domain.generation,
           domain.requested_by, domain.approved_by, domain.reason,
           domain.change_reference, domain.acceptance_digest,
           domain.changed_at
    FROM platform.domain_cutovers AS domain
    WHERE domain.scope_kind = 'tenant'
      AND domain.scope_id = p_tenant_id
      AND domain.domain = 'tenancy'
      AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_groups AS authority
          WHERE authority.scope_kind = domain.scope_kind
            AND authority.scope_id = domain.scope_id
            AND authority.group_id = domain.domain
      );
    INSERT INTO platform.cutover_group_members (
        scope_kind, scope_id, group_id, ordinal, domain
    )
    SELECT authority.scope_kind, authority.scope_id,
           authority.group_id, 1, authority.group_id
    FROM platform.cutover_groups AS authority
    WHERE authority.scope_kind = 'tenant'
      AND authority.scope_id = p_tenant_id
      AND authority.group_id = 'tenancy'
      AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_group_members AS member
          WHERE member.scope_kind = authority.scope_kind
            AND member.scope_id = authority.scope_id
            AND member.group_id = authority.group_id
            AND member.domain = authority.group_id
      );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_ensure_tenancy_authority_v1(uuid)
    FROM PUBLIC;
