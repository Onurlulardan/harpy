-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_get_signing_bundle_v1(
    p_environment_id uuid, p_generation bigint
)
RETURNS TABLE(
    environment_id uuid, generation bigint, bundle_sha256 text,
    bundle_bytes bytea, created_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT b.environment_id, b.generation, b.bundle_sha256,
           b.bundle_bytes, b.created_at
      FROM platform.platformops_recovery_signing_bundles b
     WHERE b.environment_id = p_environment_id AND b.generation = p_generation
    ;
END;
$function$;

CREATE FUNCTION platform.platformops_claim_recovery_set_v1(
    p_recovery_set_id uuid, p_claimed_at timestamptz
)
RETURNS TABLE(
    recovery_set_id uuid, environment_id uuid, sequence bigint,
    primary_binding_id uuid, tenant_set_sha256 text, owner_binding_id uuid,
    owner_binding_generation bigint, database_lineage_id uuid,
    recovery_signing_generation bigint, recovery_trust_bundle_sha256 text
)
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    UPDATE platform.platformops_recovery_sets r
       SET state = 'holding', updated_at = p_claimed_at
     WHERE r.recovery_set_id = p_recovery_set_id AND r.state = 'requested'
    RETURNING r.recovery_set_id, r.environment_id, r.sequence,
              r.primary_binding_id, r.tenant_set_sha256, r.owner_binding_id,
              r.owner_binding_generation, r.database_lineage_id,
              r.recovery_signing_generation, r.recovery_trust_bundle_sha256
    ;
END;
$function$;

CREATE FUNCTION platform.platformops_get_recovery_set_v1(p_recovery_set_id uuid)
RETURNS TABLE(
    recovery_set_id uuid, operation_id uuid, environment_id uuid,
    sequence bigint, state text, manifest_sha256 text, manifest_bytes bytea,
    legal_hold boolean, database_system_id uuid, recovery_signing_generation bigint,
    recovery_trust_bundle_sha256 text,
    protected_at timestamptz, created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT r.recovery_set_id, r.operation_id, r.environment_id,
           r.sequence, r.state, r.manifest_sha256, r.manifest_bytes,
           r.legal_hold, r.database_system_id, r.recovery_signing_generation,
           r.recovery_trust_bundle_sha256,
           r.protected_at, r.created_at, r.updated_at
      FROM platform.platformops_recovery_sets r
     WHERE r.recovery_set_id = p_recovery_set_id
    ;
END;
$function$;

CREATE FUNCTION platform.platformops_list_retention_candidates_v1(p_environment_id uuid)
RETURNS TABLE(
    recovery_set_id uuid, manifest_sha256 text,
    protected_at timestamptz, legal_hold boolean
)
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT r.recovery_set_id, r.manifest_sha256, r.protected_at, r.legal_hold
      FROM platform.platformops_recovery_sets r
     WHERE r.environment_id = p_environment_id AND r.state = 'protected'
     ORDER BY r.protected_at DESC, r.recovery_set_id DESC
    ;
END;
$function$;

CREATE FUNCTION platform.platformops_claim_restore_v1(
    p_operation_id uuid, p_claimed_at timestamptz
)
RETURNS TABLE(
    operation_id uuid, recovery_set_id uuid, environment_id uuid,
    target_disposition text, plan_sha256 text, plan_bytes bytea
)
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    UPDATE platform.platformops_restore_operations r
       SET state = 'draining', updated_at = p_claimed_at
     WHERE r.operation_id = p_operation_id AND r.state = 'plan_ready'
    RETURNING r.operation_id, r.recovery_set_id, r.environment_id,
              r.target_disposition, r.plan_sha256, r.plan_bytes
    ;
END;
$function$;

CREATE FUNCTION platform.platformops_get_restore_v1(p_operation_id uuid)
RETURNS TABLE(
    operation_id uuid, recovery_set_id uuid, environment_id uuid,
    target_disposition text, state text, plan_sha256 text,
    receipt_sha256 text, new_database_system_id uuid,
    created_at timestamptz, updated_at timestamptz, completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT r.operation_id, r.recovery_set_id, r.environment_id,
           r.target_disposition, r.state, r.plan_sha256,
           r.receipt_sha256, r.new_database_system_id,
           r.created_at, r.updated_at, r.completed_at
      FROM platform.platformops_restore_operations r
     WHERE r.operation_id = p_operation_id
    ;
END;
$function$;
-- +goose StatementEnd
