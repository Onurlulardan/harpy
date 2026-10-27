-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_resolve_current_global_authority_v1(
    p_authority_group_id text
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_authority_id uuid;
BEGIN
    IF p_authority_group_id IS DISTINCT FROM 'billing'
       OR SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing global authority resolver session is invalid';
    END IF;
    SELECT authority.authority_id INTO STRICT v_authority_id
    FROM platform.billing_global_authorities AS authority
    JOIN platform.environment_identity AS identity
      ON identity.environment_id = authority.scope_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = authority.scope_kind
     AND group_row.scope_id = authority.scope_id
     AND group_row.group_id = authority.authority_group_id
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = authority.scope_kind
     AND domain_row.scope_id = authority.scope_id
     AND domain_row.domain = authority.authority_domain
    WHERE authority.scope_kind = 'global'
      AND authority.authority_group_id = p_authority_group_id
      AND authority.authority_domain = 'billing'
      AND authority.authority_group_generation = group_row.generation
      AND authority.authority_domain_generation = domain_row.generation
      AND group_row.writer = 'v2' AND group_row.state = 'v2'
      AND domain_row.writer = 'v2' AND domain_row.state = 'v2';
    RETURN v_authority_id;
EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
    RAISE EXCEPTION USING ERRCODE = '42501',
        MESSAGE = 'billing global authority is not current';
END;
$function$;

CREATE FUNCTION platform.billing_effective_subscription_price_snapshot_read_v1(
    p_authority_id uuid, p_tenant_id uuid,
    p_subscription_id uuid, p_at timestamptz
)
RETURNS TABLE (
    effective_minor_units bigint,
    effective_currency text,
    effective_registry_version bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_state text;
    v_price_version_id uuid;
    v_applied_kind text;
    v_applied_currency text;
    v_applied_registry bigint;
    v_applied_fixed bigint;
    v_applied_basis_points integer;
    v_quotient bigint;
    v_remainder bigint;
    v_discount bigint;
BEGIN
    IF p_authority_id IS NULL OR p_tenant_id IS NULL
       OR p_subscription_id IS NULL OR p_at IS NULL
       OR SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_authority_id IS DISTINCT FROM
          platform.billing_resolve_current_tenant_authority_v1(p_tenant_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing effective snapshot price authority is invalid';
    END IF;
    SELECT subscription.state, subscription.price_version_id,
           price.minor_units, pg_catalog.btrim(price.currency::text),
           price.currency_registry_version
    INTO STRICT v_state, v_price_version_id, effective_minor_units,
        effective_currency, effective_registry_version
    FROM platform.billing_subscriptions AS subscription
    JOIN platform.billing_price_versions AS price
      ON price.id = subscription.price_version_id
     AND price.offering_version_id = subscription.offering_version_id
    WHERE subscription.authority_id = p_authority_id
      AND subscription.tenant_id = p_tenant_id
      AND subscription.id = p_subscription_id;
    IF v_state = 'trialing' THEN
        effective_minor_units := 0;
        RETURN NEXT;
        RETURN;
    END IF;
    FOR v_applied_kind, v_applied_currency, v_applied_registry,
        v_applied_fixed, v_applied_basis_points IN
        SELECT redemption.applied_kind,
               pg_catalog.btrim(redemption.applied_currency::text),
               redemption.applied_currency_registry_version,
               redemption.applied_fixed_minor_units,
               redemption.applied_basis_points
        FROM platform.billing_discount_redemptions AS redemption
        WHERE redemption.authority_id = p_authority_id
          AND redemption.tenant_id = p_tenant_id
          AND redemption.subscription_id = p_subscription_id
          AND redemption.price_version_id = v_price_version_id
          AND redemption.redeemed_at <= p_at
          AND (redemption.relationship_ended_at IS NULL
               OR redemption.relationship_ended_at > p_at)
        ORDER BY redemption.applied_priority, redemption.discount_version_id
    LOOP
        IF v_applied_kind = 'fixed' THEN
            IF v_applied_currency IS DISTINCT FROM effective_currency
               OR v_applied_registry IS DISTINCT FROM effective_registry_version THEN
                RAISE EXCEPTION USING ERRCODE = '23514',
                    MESSAGE = 'billing discount currency invariant failed';
            END IF;
            v_discount := v_applied_fixed;
        ELSE
            SELECT pg_catalog.floor(effective_minor_units::numeric
                       * v_applied_basis_points::numeric / 10000)::bigint,
                   pg_catalog.mod(effective_minor_units::numeric
                       * v_applied_basis_points::numeric, 10000)::bigint
            INTO v_quotient, v_remainder;
            IF v_remainder > 5000
               OR (v_remainder = 5000 AND pg_catalog.mod(v_quotient, 2) = 1) THEN
                v_quotient := v_quotient + 1;
            END IF;
            v_discount := v_quotient;
        END IF;
        effective_minor_units := GREATEST(
            0, effective_minor_units - v_discount
        );
    END LOOP;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.billing_resolve_current_global_authority_v1(text) FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.billing_effective_subscription_price_snapshot_read_v1(
        uuid,uuid,uuid,timestamptz
    ) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.billing_resolve_current_global_authority_v1(text)
    TO harpy_runtime;
GRANT EXECUTE ON FUNCTION
    platform.billing_effective_subscription_price_snapshot_read_v1(
        uuid,uuid,uuid,timestamptz
    ) TO harpy_runtime;
