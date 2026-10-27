-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_restore_binding_v1(
    p_environment_id uuid
)
RETURNS TABLE (
    license_id uuid, licensee_organization_id uuid, environment_id uuid,
    delivery_kind text, issuer text, sequence bigint,
    payload_sha256 text, authority_revision bigint, database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_count integer;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR pg_catalog.current_setting('application_name', true)
            <> 'forgecrud-v2:core-operator-license-reconcile'
       OR p_environment_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing restored license binding denied';
    END IF;
    IF NOT EXISTS (
        SELECT 1
          FROM platform.tenancy_deployment_policy AS policy
          JOIN platform.tenancy_tenant_bindings AS binding
            ON binding.environment_id = policy.environment_id
           AND binding.purpose = 'onpremise_primary'
         WHERE policy.singleton
           AND policy.environment_id = p_environment_id
           AND policy.deployment_mode = 'onpremise'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing restored license requires onpremise primary';
    END IF;
    SELECT pg_catalog.count(*) INTO v_count
      FROM platform.billing_license_heads AS head
      JOIN platform.tenancy_deployment_policy AS policy
        ON policy.environment_id = head.environment_id
       AND policy.singleton
       AND policy.deployment_mode = 'onpremise'
     WHERE head.environment_id = p_environment_id;
    IF v_count = 0 THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'billing restored license binding is unavailable';
    END IF;
    IF v_count <> 1 THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing restored license binding is ambiguous';
    END IF;
    RETURN QUERY
    SELECT head.license_id, snapshot.licensee_organization_id,
        head.environment_id, snapshot.delivery_kind, snapshot.issuer,
        head.sequence, head.payload_sha256, head.revision,
        pg_catalog.clock_timestamp()
      FROM platform.billing_license_heads AS head
      JOIN platform.billing_license_snapshots AS snapshot
        ON snapshot.snapshot_id = head.snapshot_id
      JOIN platform.tenancy_deployment_policy AS policy
        ON policy.environment_id = head.environment_id
       AND policy.singleton
       AND policy.deployment_mode = 'onpremise'
     WHERE head.environment_id = p_environment_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_restore_binding_v1(uuid)
FROM PUBLIC;
