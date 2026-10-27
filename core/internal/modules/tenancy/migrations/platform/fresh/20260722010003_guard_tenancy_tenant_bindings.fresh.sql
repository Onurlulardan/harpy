-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_guard_tenant_binding_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
DECLARE
    policy_mode text;
BEGIN
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'harpy:v2:tenancy-binding:' || NEW.environment_id::text,
            0
        )
    );
    SELECT policy.deployment_mode
    INTO policy_mode
    FROM platform.tenancy_deployment_policy AS policy
    WHERE policy.environment_id = NEW.environment_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'Persisted tenancy deployment policy is required';
    END IF;

    IF policy_mode = 'onpremise' THEN
        IF NEW.purpose IS DISTINCT FROM 'onpremise_primary' THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'On-premise binding purpose is invalid';
        END IF;
        IF EXISTS (
            SELECT 1 FROM platform.tenancy_tenant_bindings AS binding
            WHERE binding.environment_id = NEW.environment_id
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'On-premise tenant capacity is exhausted';
        END IF;
    ELSIF policy_mode = 'saas' THEN
        IF NEW.purpose = 'saas_control' AND EXISTS (
            SELECT 1 FROM platform.tenancy_tenant_bindings AS binding
            WHERE binding.environment_id = NEW.environment_id
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'SaaS control binding already exists';
        ELSIF NEW.purpose = 'saas_customer' AND NOT EXISTS (
            SELECT 1 FROM platform.tenancy_tenant_bindings AS binding
            WHERE binding.environment_id = NEW.environment_id
              AND binding.purpose = 'saas_control'
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'SaaS control binding is required before customers';
        ELSIF NEW.purpose NOT IN ('saas_control', 'saas_customer') THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'SaaS binding purpose is invalid';
        END IF;
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER tenancy_tenant_bindings_guard_insert
BEFORE INSERT ON platform.tenancy_tenant_bindings
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_guard_tenant_binding_v1();

CREATE TRIGGER tenancy_tenant_bindings_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenancy_tenant_bindings
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_mutation();
