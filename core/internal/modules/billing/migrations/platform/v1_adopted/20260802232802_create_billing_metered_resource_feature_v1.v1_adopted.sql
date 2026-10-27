-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_metered_resource_feature_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_entitlement_key text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_assignment platform.billing_installation_base_assignments%ROWTYPE;
    v_enabled boolean;
BEGIN
    PERFORM platform.billing_assert_tenant_authority_v1(p_authority_id);
    IF p_entitlement_key NOT IN (
        'api.builder.enabled','analytics.dashboard_builder.enabled'
    ) OR NOT EXISTS (
        SELECT 1
        FROM platform.billing_tenant_authorities AS authority_row
        WHERE authority_row.authority_id = p_authority_id
          AND authority_row.tenant_id = p_tenant_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing metered resource feature input is invalid';
    END IF;

    SELECT assignment.*
    INTO STRICT v_assignment
    FROM platform.billing_installation_base_assignments AS assignment
    WHERE assignment.authority_id = p_authority_id
      AND assignment.tenant_id = p_tenant_id;

    IF v_assignment.deployment_mode = 'onpremise' THEN
        SELECT runtime.boolean_value
        INTO STRICT v_enabled
        FROM platform.billing_license_runtime_entitlements_v1(
            v_assignment.environment_id
        ) AS runtime
        WHERE runtime.entitlement_key = p_entitlement_key
          AND runtime.value_type = 'boolean'
          AND runtime.boolean_value IS NOT NULL;
        RETURN v_enabled;
    END IF;
    IF v_assignment.deployment_mode <> 'saas' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing deployment mode is unavailable';
    END IF;

    SELECT CASE
               WHEN override_row.id IS NOT NULL THEN override_row.boolean_value
               WHEN purchased.entitlement_definition_id IS NOT NULL
                   THEN purchased.boolean_value
               ELSE base_grant.boolean_value
           END
    INTO STRICT v_enabled
    FROM platform.billing_entitlement_definitions AS definition
    JOIN LATERAL (
        SELECT base_assignment.offering_version_id
        FROM platform.billing_tenant_base_offerings AS base_assignment
        WHERE base_assignment.authority_id = p_authority_id
          AND base_assignment.tenant_id = p_tenant_id
          AND base_assignment.effective_start <= pg_catalog.statement_timestamp()
          AND (
              base_assignment.effective_end IS NULL
              OR base_assignment.effective_end > pg_catalog.statement_timestamp()
          )
        ORDER BY base_assignment.effective_start DESC
        LIMIT 1
    ) AS active_base ON true
    LEFT JOIN platform.billing_offering_entitlements AS base_grant
      ON base_grant.offering_version_id = active_base.offering_version_id
     AND base_grant.entitlement_definition_id = definition.id
    LEFT JOIN LATERAL (
        SELECT subscription.id, subscription.offering_version_id
        FROM platform.billing_subscriptions AS subscription
        WHERE subscription.authority_id = p_authority_id
          AND subscription.tenant_id = p_tenant_id
          AND (
              (subscription.state = 'trialing'
               AND pg_catalog.statement_timestamp() < subscription.trial_ends_at)
              OR (subscription.state = 'active'
               AND pg_catalog.statement_timestamp() < subscription.period_end)
              OR (subscription.state = 'past_due'
               AND pg_catalog.statement_timestamp() < subscription.grace_ends_at)
          )
        LIMIT 1
    ) AS active_subscription ON true
    LEFT JOIN platform.billing_offering_entitlements AS purchased
      ON purchased.offering_version_id = active_subscription.offering_version_id
     AND purchased.entitlement_definition_id = definition.id
    LEFT JOIN LATERAL (
        SELECT candidate.*
        FROM platform.billing_subscription_overrides AS candidate
        WHERE candidate.authority_id = p_authority_id
          AND candidate.tenant_id = p_tenant_id
          AND candidate.subscription_id = active_subscription.id
          AND candidate.entitlement_definition_id = definition.id
          AND candidate.effective_start <= pg_catalog.statement_timestamp()
          AND (
              candidate.effective_end IS NULL
              OR candidate.effective_end > pg_catalog.statement_timestamp()
          )
        ORDER BY candidate.effective_start DESC
        LIMIT 1
    ) AS override_row ON true
    WHERE definition.entitlement_key = p_entitlement_key
      AND definition.value_type = 'boolean'
      AND (
          override_row.id IS NOT NULL
          OR purchased.entitlement_definition_id IS NOT NULL
          OR base_grant.entitlement_definition_id IS NOT NULL
      );

    IF v_enabled IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing metered resource feature is unavailable';
    END IF;
    RETURN v_enabled;
EXCEPTION
    WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing metered resource feature is unavailable';
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_metered_resource_feature_v1(
    uuid,uuid,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_metered_resource_feature_v1(
    uuid,uuid,text
) TO harpy_runtime;
-- +goose StatementEnd
