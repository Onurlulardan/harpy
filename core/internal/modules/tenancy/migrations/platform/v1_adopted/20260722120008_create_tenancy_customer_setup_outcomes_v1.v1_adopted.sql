-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_setup_delivery_outcome_v1(
    p_tenant_id uuid, p_artifact_id uuid, p_outcome text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_current text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'harpy-v2:core-worker'
       OR p_tenant_id IS NULL OR p_artifact_id IS NULL
       OR p_outcome NOT IN ('delivered','permanent_failure','outcome_unknown') THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='customer setup delivery outcome denied';
    END IF;
    SELECT outcome.setup_delivery_state INTO v_current
    FROM platform.tenancy_customer_provisioning_outcomes AS outcome
    JOIN platform.tenancy_onboarding_operations AS operation
      ON operation.operation_id=outcome.operation_id
    WHERE operation.tenant_id=p_tenant_id
      AND operation.operation_kind='saas_customer_tenant_provision'
      AND operation.deployment_mode='saas'
      AND outcome.setup_artifact_id=p_artifact_id
    FOR UPDATE OF outcome;
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    IF v_current=p_outcome THEN
        RETURN true;
    END IF;
    IF v_current NOT IN ('pending','reissued') THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='customer setup delivery outcome conflicts';
    END IF;
    UPDATE platform.tenancy_customer_provisioning_outcomes
    SET setup_delivery_state=p_outcome,revision=revision+1,
        updated_at=pg_catalog.statement_timestamp()
    WHERE setup_artifact_id=p_artifact_id;
    RETURN false;
END;
$function$;
CREATE FUNCTION platform.tenancy_customer_setup_accepted_v1(
    p_operation_id uuid, p_tenant_id uuid, p_artifact_id uuid
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_current text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'harpy-v2:core-api'
       OR p_operation_id IS NULL OR p_tenant_id IS NULL
       OR p_artifact_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='customer setup acceptance denied';
    END IF;
    SELECT outcome.setup_delivery_state INTO v_current
    FROM platform.tenancy_customer_provisioning_outcomes AS outcome
    JOIN platform.tenancy_onboarding_operations AS operation
      ON operation.operation_id=outcome.operation_id
    WHERE operation.operation_id=p_operation_id
      AND operation.tenant_id=p_tenant_id
      AND operation.operation_kind='saas_customer_tenant_provision'
      AND operation.deployment_mode='saas' AND operation.state='succeeded'
      AND outcome.setup_artifact_id=p_artifact_id
    FOR UPDATE OF outcome;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='customer setup acceptance conflicts';
    END IF;
    IF v_current='accepted' THEN
        RETURN true;
    END IF;
    IF v_current<>'delivered' THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='customer setup must be delivered before acceptance';
    END IF;
    UPDATE platform.tenancy_customer_provisioning_outcomes
    SET setup_delivery_state='accepted',revision=revision+1,
        updated_at=pg_catalog.statement_timestamp()
    WHERE operation_id=p_operation_id;
    RETURN false;
END;
$function$;
REVOKE ALL ON FUNCTION platform.tenancy_customer_setup_delivery_outcome_v1(
    uuid,uuid,text
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_customer_setup_accepted_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_setup_delivery_outcome_v1(
    uuid,uuid,text
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_setup_accepted_v1(
    uuid,uuid,uuid
) TO harpy_runtime;
-- +goose StatementEnd
