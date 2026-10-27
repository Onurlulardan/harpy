-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_v1_adopt_environment_v1(
    p_operation_id uuid,p_environment_id uuid,p_deployment_mode text,
    p_primary_tenant_id uuid,p_inventory_sha256 text,p_plan_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.tenancy_v1_adoption_receipts%ROWTYPE;
    v_tenant_count integer;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name',true)
            !~ '^harpy-cutover/[0-9a-f-]{36}$'
       OR p_operation_id IS NULL OR p_environment_id IS NULL
       OR p_primary_tenant_id IS NULL
       OR p_deployment_mode NOT IN ('onpremise','saas')
       OR p_inventory_sha256 !~ '^[0-9a-f]{64}$'
       OR p_plan_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='V1 tenancy adoption denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended('v1-tenancy-adopt|'||p_environment_id::text,0));
    SELECT * INTO v_existing FROM platform.tenancy_v1_adoption_receipts
    WHERE environment_id=p_environment_id FOR SHARE;
    IF FOUND THEN
        IF v_existing.operation_id IS DISTINCT FROM p_operation_id
           OR v_existing.deployment_mode IS DISTINCT FROM p_deployment_mode
           OR v_existing.primary_tenant_id IS DISTINCT FROM p_primary_tenant_id
           OR v_existing.inventory_sha256 IS DISTINCT FROM pg_catalog.decode(p_inventory_sha256,'hex')
           OR v_existing.plan_sha256 IS DISTINCT FROM pg_catalog.decode(p_plan_sha256,'hex') THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='V1 tenancy adoption replay conflicts';
        END IF;
        RETURN true;
    END IF;
    SELECT pg_catalog.count(*)::integer INTO v_tenant_count
    FROM platform.tenants AS tenant WHERE tenant.deleted_at IS NULL;
    IF (p_deployment_mode='onpremise' AND (v_tenant_count<>1 OR NOT EXISTS (
            SELECT 1 FROM platform.tenants AS tenant
            WHERE tenant.id=p_primary_tenant_id AND tenant.deleted_at IS NULL)))
       OR (p_deployment_mode='saas' AND (v_tenant_count<1 OR NOT EXISTS (
            SELECT 1 FROM platform.tenants AS tenant
            WHERE tenant.id=p_primary_tenant_id AND tenant.deleted_at IS NULL))) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 tenant inventory is ambiguous';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.tenancy_deployment_policy)
       OR EXISTS (SELECT 1 FROM platform.tenancy_tenant_bindings) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 tenancy authority already exists without adoption receipt';
    END IF;
    INSERT INTO platform.tenancy_deployment_policy(
        environment_id,deployment_mode,maximum_customer_tenants,created_at
    ) VALUES (p_environment_id,p_deployment_mode,
        CASE WHEN p_deployment_mode='onpremise' THEN 0 ELSE NULL END,v_now);
    INSERT INTO platform.tenancy_tenant_bindings(
        environment_id,tenant_id,purpose,created_at
    ) VALUES (p_environment_id,p_primary_tenant_id,
        CASE WHEN p_deployment_mode='onpremise'
            THEN 'onpremise_primary' ELSE 'saas_control' END,v_now);
    IF p_deployment_mode='saas' THEN
        INSERT INTO platform.tenancy_tenant_bindings(
            environment_id,tenant_id,purpose,created_at
        ) SELECT p_environment_id,tenant.id,'saas_customer',v_now
          FROM platform.tenants AS tenant
         WHERE tenant.deleted_at IS NULL AND tenant.id<>p_primary_tenant_id
         ORDER BY tenant.id;
    END IF;
    INSERT INTO platform.tenancy_v1_adoption_receipts(
        operation_id,environment_id,deployment_mode,primary_tenant_id,
        adopted_tenant_count,inventory_sha256,plan_sha256,adopted_at,recorded_by
    ) VALUES (p_operation_id,p_environment_id,p_deployment_mode,p_primary_tenant_id,
        v_tenant_count,pg_catalog.decode(p_inventory_sha256,'hex'),
        pg_catalog.decode(p_plan_sha256,'hex'),v_now,SESSION_USER);
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_v1_adopt_environment_v1(
    uuid,uuid,text,uuid,text,text
) FROM PUBLIC;
