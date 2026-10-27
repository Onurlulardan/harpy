-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_provision_claim_v1(
    p_attempt_id uuid,p_lease_owner text,p_lease_token_sha256 text,
    p_lease_seconds integer
)
RETURNS TABLE (
    operation_id uuid,environment_id uuid,operation_kind text,
    idempotency_key text,request_sha256 text,manifest_sha256 text,
    step_plan_sha256 text,deployment_mode text,tenant_id uuid,
    schema_name text,binding_purpose text,first_user_id uuid,
    root_organization_id uuid,protected_admin_binding_id uuid,
    instance_owner_binding_id uuid,offering_version_id uuid,
    offering_revision bigint,offering_sha256 text,
    offering_effective_at timestamptz,offering_source_sequence bigint,
    offering_source_sha256 text,recovery_signing_generation bigint,
    recovery_trust_bundle_sha256 text,registry_binding_id uuid,
    authority_scope_kind text,authority_scope_id uuid,
    authority_generation bigint,authority_previous_head_sha256 text,
    authority_state_sha256 text,authority_proposal_sha256 text,
    created_at timestamptz,customer_name text,tenant_slug text,
    first_admin_display_name text,first_admin_email text,
    state text,revision bigint,lease_until timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.tenancy_onboarding_operations%ROWTYPE;
    v_original_state text;
    v_revision bigint;
    v_lease_until timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR p_attempt_id IS NULL OR pg_catalog.octet_length(p_lease_owner) NOT BETWEEN 1 AND 128
       OR p_lease_token_sha256 !~ '^[0-9a-f]{64}$'
       OR p_lease_seconds NOT BETWEEN 1 AND 300 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='customer provisioning worker claim is denied';
    END IF;
    SELECT operation.* INTO v_operation
    FROM platform.tenancy_onboarding_operations AS operation
    LEFT JOIN platform.tenancy_onboarding_leases AS lease
      ON lease.operation_id=operation.operation_id
    WHERE operation.operation_kind='saas_customer_tenant_provision'
      AND operation.deployment_mode='saas'
      AND operation.state IN (
          'migration_ready','authority_pending','authority_acknowledged'
      )
      AND (lease.operation_id IS NULL
           OR lease.lease_until <= pg_catalog.statement_timestamp())
    ORDER BY operation.created_at,operation.operation_id
    LIMIT 1 FOR UPDATE OF operation SKIP LOCKED;
    IF NOT FOUND THEN
        RETURN;
    END IF;
    v_original_state:=v_operation.state;
    SELECT claimed.revision,claimed.lease_until
    INTO STRICT v_revision,v_lease_until
    FROM platform.tenancy_onboarding_claim_v2(
        v_operation.operation_id,p_attempt_id,p_lease_owner,
        p_lease_token_sha256,v_operation.revision,p_lease_seconds
    ) AS claimed;
    IF v_original_state='migration_ready' THEN
        UPDATE platform.tenancy_onboarding_operations AS operation
        SET state='authority_pending',revision=v_revision+1,
            updated_at=pg_catalog.statement_timestamp()
        WHERE operation.operation_id=v_operation.operation_id
          AND operation.state='migration_ready' AND operation.revision=v_revision;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='40001',
                MESSAGE='customer authority transition conflicts';
        END IF;
        v_revision:=v_revision+1;
        UPDATE platform.tenancy_onboarding_leases AS lease
        SET revision=v_revision,heartbeat_at=pg_catalog.statement_timestamp()
        WHERE lease.operation_id=v_operation.operation_id
          AND lease.attempt_id=p_attempt_id;
    END IF;
    RETURN QUERY
    SELECT operation.operation_id,operation.environment_id,
        operation.operation_kind,operation.idempotency_key,
        operation.request_sha256,operation.manifest_sha256,
        operation.step_plan_sha256,operation.deployment_mode,
        operation.tenant_id,operation.schema_name::text,
        operation.binding_purpose,operation.first_user_id,
        operation.root_organization_id,operation.protected_admin_binding_id,
        operation.instance_owner_binding_id,operation.offering_version_id,
        operation.offering_revision,operation.offering_sha256,
        operation.offering_effective_at,operation.offering_source_sequence,
        operation.offering_source_sha256,operation.recovery_signing_generation,
        operation.recovery_trust_bundle_sha256,operation.registry_binding_id,
        operation.authority_scope_kind,operation.authority_scope_id,
        operation.authority_generation,operation.authority_previous_head_sha256,
        operation.authority_state_sha256,operation.authority_proposal_sha256,
        operation.created_at,request.customer_name,request.tenant_slug,
        request.first_admin_display_name,request.first_admin_email,
        operation.state,operation.revision,v_lease_until
    FROM platform.tenancy_onboarding_operations AS operation
    JOIN platform.tenancy_customer_provisioning_requests AS request
      ON request.operation_id=operation.operation_id
    WHERE operation.operation_id=v_operation.operation_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_provision_claim_v1(
    uuid,text,text,integer
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_provision_claim_v1(
    uuid,text,text,integer
) TO forgecrud_runtime;

