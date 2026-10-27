-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_assert_cutover_shadow_authority_v1(
    p_scope_kind text, p_scope_id uuid, p_group_id text,
    p_group_generation bigint, p_domain text, p_cutover_generation bigint
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF p_scope_kind IS DISTINCT FROM 'tenant' OR p_scope_id IS NULL OR
       p_group_id IS DISTINCT FROM 'documents' OR p_group_generation < 1 OR
       p_domain IS DISTINCT FROM 'documents' OR p_cutover_generation < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'cutover shadow authority input is invalid';
    END IF;
    PERFORM 1
    FROM platform.cutover_job_bindings AS binding
    JOIN platform.cutover_operations AS operation
      ON operation.operation_id=binding.operation_id
     AND operation.run_id=binding.cutover_run_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind=operation.scope_kind
     AND group_row.scope_id=operation.scope_id
     AND group_row.group_id=operation.group_id
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind=operation.scope_kind
     AND domain_row.scope_id=operation.scope_id
     AND domain_row.domain=operation.domain
    JOIN platform.cutover_group_members AS member
      ON member.scope_kind=operation.scope_kind
     AND member.scope_id=operation.scope_id
     AND member.group_id=operation.group_id
     AND member.domain=operation.domain
    WHERE binding.tenant_id=p_scope_id
      AND binding.group_id=p_group_id
      AND binding.group_generation=p_group_generation
      AND binding.domain=p_domain
      AND binding.cutover_generation=p_cutover_generation
      AND operation.scope_kind=p_scope_kind
      AND operation.scope_id=p_scope_id
      AND operation.group_id=p_group_id
      AND operation.domain=p_domain
      AND operation.cutover_generation=p_cutover_generation
      AND operation.lease_manifest_sha256=binding.lease_manifest_sha256
      AND operation.state IN ('copying','shadow')
      AND operation.lease_expires_at>pg_catalog.clock_timestamp()
      AND group_row.state IN ('v1','shadow') AND group_row.writer='v1'
      AND group_row.generation=p_group_generation
      AND domain_row.state IN ('v1','shadow') AND domain_row.writer='v1'
      AND domain_row.generation=p_cutover_generation
    LIMIT 1
    FOR SHARE OF binding,operation,group_row,domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5601',
            MESSAGE = 'cutover shadow authority is stale';
    END IF;
END;
$function$;
CREATE FUNCTION platform.platform_async_assert_worker_authority_v1(
    p_scope_kind text, p_scope_id uuid, p_group_id text,
    p_group_generation bigint, p_domain text, p_cutover_generation bigint,
    p_mode text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF p_mode='v2' THEN
        PERFORM platform.platform_async_assert_current_authority_v1(
            p_scope_kind,p_scope_id,p_group_id,p_group_generation,
            p_domain,p_cutover_generation);
    ELSIF p_mode='cutover_shadow' THEN
        PERFORM platform.platform_async_assert_cutover_shadow_authority_v1(
            p_scope_kind,p_scope_id,p_group_id,p_group_generation,
            p_domain,p_cutover_generation);
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'job authority mode is invalid';
    END IF;
END;
$function$;
CREATE FUNCTION platform.platform_async_cutover_shadow_operation_active_v1(
    p_operation_id uuid, p_run_id uuid, p_scope_kind text, p_scope_id uuid,
    p_group_id text, p_domain text, p_cutover_generation bigint,
    p_manifest_sha256 text, p_observed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_active boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM platform.cutover_operations AS operation
        WHERE operation.operation_id=p_operation_id
          AND operation.run_id=p_run_id
          AND operation.scope_kind=p_scope_kind
          AND operation.scope_id=p_scope_id
          AND operation.group_id=p_group_id
          AND operation.domain=p_domain
          AND operation.cutover_generation=p_cutover_generation
          AND operation.lease_manifest_sha256=p_manifest_sha256
          AND operation.state IN ('copying','shadow')
          AND operation.lease_expires_at>p_observed_at
    ) INTO v_active;
    RETURN v_active;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_assert_cutover_shadow_authority_v1(
    text,uuid,text,bigint,text,bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_async_assert_worker_authority_v1(
    text,uuid,text,bigint,text,bigint,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION
platform.platform_async_cutover_shadow_operation_active_v1(
    uuid,uuid,text,uuid,text,text,bigint,text,timestamptz
) FROM PUBLIC;
