-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_refresh_revenue_projection_v1(
    p_subscription_id uuid
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    current_authority uuid;
BEGIN
    SELECT authority.authority_id
    INTO current_authority
    FROM platform.billing_global_authorities AS authority
    JOIN platform.environment_identity AS identity
      ON identity.environment_id = authority.scope_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = authority.scope_kind
     AND group_row.scope_id = authority.scope_id
     AND group_row.group_id = authority.authority_group_id
     AND group_row.generation = authority.authority_group_generation
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = authority.scope_kind
     AND domain_row.scope_id = authority.scope_id
     AND domain_row.domain = authority.authority_domain
     AND domain_row.generation = authority.authority_domain_generation
    WHERE authority.scope_kind = 'global'
      AND authority.authority_group_id = 'billing'
      AND authority.authority_domain = 'billing'
    FOR SHARE OF authority, group_row, domain_row;

    IF current_authority IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing revenue global authority is not current';
    END IF;
    PERFORM platform.billing_assert_global_authority_v1(current_authority);

    DELETE FROM platform.billing_revenue_projection AS projection
    WHERE projection.subscription_id = p_subscription_id
      AND NOT EXISTS (
        SELECT 1
        FROM platform.billing_subscriptions AS subscription
        JOIN platform.billing_price_versions AS price
          ON price.id = subscription.price_version_id
         AND price.offering_version_id = subscription.offering_version_id
        WHERE subscription.id = p_subscription_id
          AND subscription.state IN ('active', 'past_due')
          AND price.cadence IN ('monthly', 'yearly')
      );

    INSERT INTO platform.billing_revenue_projection (
        id, authority_id, subscription_id, subscription_revision,
        offering_version_id, price_version_id, price_source_sha256,
        state, cadence, currency, currency_registry_version,
        source_minor_units, monthly_minor_units, annual_minor_units,
        refreshed_at
    )
    SELECT
        platform.billing_revenue_projection_id_v1(subscription.id),
        current_authority, subscription.id, subscription.revision,
        subscription.offering_version_id, subscription.price_version_id,
        pg_catalog.sha256(pg_catalog.convert_to(pg_catalog.concat_ws(
            pg_catalog.chr(31), price.id::text,
            subscription.offering_version_id::text, price.cadence,
            price.minor_units::text, price.currency::text,
            price.currency_registry_version::text, price.money_policy_version
        ), 'UTF8')),
        subscription.state, price.cadence, price.currency,
        price.currency_registry_version, price.minor_units,
        platform.billing_revenue_monthly_v1(
            price.minor_units, price.cadence
        ),
        CASE WHEN price.cadence = 'monthly'
            THEN price.minor_units * 12
            ELSE price.minor_units
        END,
        pg_catalog.clock_timestamp()
    FROM platform.billing_subscriptions AS subscription
    JOIN platform.billing_price_versions AS price
      ON price.id = subscription.price_version_id
     AND price.offering_version_id = subscription.offering_version_id
    WHERE subscription.id = p_subscription_id
      AND subscription.state IN ('active', 'past_due')
      AND price.cadence IN ('monthly', 'yearly')
    ON CONFLICT (subscription_id) DO UPDATE SET
        authority_id = EXCLUDED.authority_id,
        subscription_revision = EXCLUDED.subscription_revision,
        offering_version_id = EXCLUDED.offering_version_id,
        price_version_id = EXCLUDED.price_version_id,
        price_source_sha256 = EXCLUDED.price_source_sha256,
        state = EXCLUDED.state,
        cadence = EXCLUDED.cadence,
        currency = EXCLUDED.currency,
        currency_registry_version = EXCLUDED.currency_registry_version,
        source_minor_units = EXCLUDED.source_minor_units,
        monthly_minor_units = EXCLUDED.monthly_minor_units,
        annual_minor_units = EXCLUDED.annual_minor_units,
        refreshed_at = EXCLUDED.refreshed_at;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.billing_refresh_revenue_projection_v1(uuid) FROM PUBLIC;
-- +goose StatementEnd
