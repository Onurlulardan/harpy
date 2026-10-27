-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_guard_tenant_authority_v2()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_mode text;
    v_state text;
    v_writer text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR TG_OP IS DISTINCT FROM 'INSERT' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics tenant authority mutation is invalid';
    END IF;
    IF SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime' THEN
        v_mode := 'runtime_v2'; v_state := 'v2'; v_writer := 'v2';
    ELSIF SESSION_USER IS NOT DISTINCT FROM 'harpy_adopter' THEN
        v_mode := 'adopter_shadow'; v_state := 'shadow'; v_writer := 'v1';
    ELSIF SESSION_USER IS NOT DISTINCT FROM 'harpy_cutover_finalizer' THEN
        v_mode := 'runtime_v2'; v_state := 'v2'; v_writer := 'v2';
    ELSE
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics tenant authority session is invalid';
    END IF;
    IF NEW.authority_mode IS DISTINCT FROM v_mode THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics tenant authority mode is invalid';
    END IF;
    PERFORM 1
    FROM platform.tenants AS tenant
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = 'tenant' AND group_row.scope_id = tenant.id
     AND group_row.group_id = 'analytics'
    JOIN platform.cutover_group_members AS member_row
      ON member_row.scope_kind = group_row.scope_kind
     AND member_row.scope_id = group_row.scope_id
     AND member_row.group_id = group_row.group_id
     AND member_row.ordinal = 1 AND member_row.domain = 'analytics'
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = member_row.scope_kind
     AND domain_row.scope_id = member_row.scope_id
     AND domain_row.domain = member_row.domain
    WHERE tenant.id = NEW.tenant_id AND tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND NEW.scope_kind = 'tenant' AND NEW.scope_id = tenant.id
      AND NEW.authority_group_id = group_row.group_id
      AND NEW.authority_domain = domain_row.domain
      AND NEW.authority_group_generation = group_row.generation
      AND NEW.authority_domain_generation = domain_row.generation
      AND group_row.state = v_state AND group_row.writer = v_writer
      AND domain_row.state = v_state AND domain_row.writer = v_writer
      AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_group_members AS extra
          WHERE extra.scope_kind = 'tenant' AND extra.scope_id = tenant.id
            AND extra.group_id = 'analytics'
            AND NOT (extra.ordinal = 1 AND extra.domain = 'analytics')
      )
    FOR SHARE OF tenant, group_row, member_row, domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6101',
            MESSAGE = 'analytics tenant authority binding is stale';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.analytics_guard_tenant_authority_v2() FROM PUBLIC;
CREATE OR REPLACE TRIGGER analytics_tenant_authorities_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.analytics_tenant_authorities
FOR EACH ROW EXECUTE FUNCTION platform.analytics_guard_tenant_authority_v2();
ALTER TABLE platform.analytics_tenant_authorities
    ENABLE ALWAYS TRIGGER analytics_tenant_authorities_guard;

CREATE FUNCTION platform.analytics_finalize_authority_successor_v1(
    p_operation_id uuid, p_run_id uuid, p_acceptance_digest text
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_authority_id uuid;
    v_tenant_id uuid;
    v_created_at timestamptz;
    v_generation bigint;
    v_run_hash text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_cutover_finalizer' OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_run_id) IS DISTINCT FROM 7 OR
       p_acceptance_digest IS NULL OR
       p_acceptance_digest !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics successor authority input is invalid';
    END IF;
    SELECT tenant.id, group_row.generation, group_row.changed_at
    INTO v_tenant_id, v_generation, v_created_at
    FROM platform.cutover_operations AS operation
    JOIN platform.tenants AS tenant
      ON operation.scope_kind = 'tenant' AND tenant.id = operation.scope_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = operation.scope_kind
     AND group_row.scope_id = operation.scope_id
     AND group_row.group_id = operation.group_id
    JOIN platform.cutover_group_members AS member_row
      ON member_row.scope_kind = group_row.scope_kind
     AND member_row.scope_id = group_row.scope_id
     AND member_row.group_id = group_row.group_id
     AND member_row.ordinal = 1 AND member_row.domain = 'analytics'
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = member_row.scope_kind
     AND domain_row.scope_id = member_row.scope_id
     AND domain_row.domain = member_row.domain
    WHERE operation.operation_id = p_operation_id AND operation.run_id = p_run_id
      AND operation.state = 'completed'
      AND operation.group_id = 'analytics' AND operation.domain = 'analytics'
      AND tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND group_row.state = 'v2' AND group_row.writer = 'v2'
      AND domain_row.state = 'v2' AND domain_row.writer = 'v2'
      AND group_row.generation = operation.cutover_generation + 1
      AND domain_row.generation = operation.cutover_generation + 1
      AND group_row.acceptance_digest = p_acceptance_digest
      AND domain_row.acceptance_digest = p_acceptance_digest
      AND group_row.requested_by = domain_row.requested_by
      AND group_row.approved_by = domain_row.approved_by
      AND group_row.approved_by IS NOT NULL
      AND group_row.reason = domain_row.reason
      AND group_row.change_reference = domain_row.change_reference
      AND group_row.changed_at = domain_row.changed_at
      AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_group_members AS extra
          WHERE extra.scope_kind = operation.scope_kind
            AND extra.scope_id = operation.scope_id
            AND extra.group_id = operation.group_id
            AND NOT (extra.ordinal = 1 AND extra.domain = 'analytics')
      )
    FOR SHARE OF operation, tenant, group_row, member_row, domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6101',
            MESSAGE = 'analytics successor authority evidence differs';
    END IF;
    v_run_hash := pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(p_run_id::text, 'UTF8')
    ), 'hex');
    v_authority_id := (
        pg_catalog.substring(p_operation_id::text, 1, 19) || '8' ||
        pg_catalog.substring(v_run_hash, 1, 3) || '-' ||
        pg_catalog.substring(v_run_hash, 4, 12)
    )::uuid;
    INSERT INTO platform.analytics_tenant_authorities (
        authority_id, tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation,
        authority_domain, authority_domain_generation, authority_mode, created_at
    ) VALUES (
        v_authority_id, v_tenant_id, 'tenant', v_tenant_id,
        'analytics', v_generation, 'analytics', v_generation,
        'runtime_v2', v_created_at
    ) ON CONFLICT DO NOTHING;
    PERFORM 1 FROM platform.analytics_tenant_authorities AS authority_row
    WHERE authority_row.authority_id = v_authority_id
      AND authority_row.tenant_id = v_tenant_id
      AND authority_row.authority_group_generation = v_generation
      AND authority_row.authority_domain_generation = v_generation
      AND authority_row.authority_mode = 'runtime_v2'
      AND authority_row.created_at IS NOT DISTINCT FROM v_created_at;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6101',
            MESSAGE = 'analytics successor authority replay differs';
    END IF;
    RETURN v_authority_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_finalize_authority_successor_v1(
    uuid,uuid,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.analytics_finalize_authority_successor_v1(
    uuid,uuid,text
) TO harpy_cutover_authority_owner;
