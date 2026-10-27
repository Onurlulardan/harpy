-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_revenue_source_digest_v1(p_authority_id uuid)
RETURNS TABLE (
    row_count bigint,
    revision_ceiling bigint,
    canonical_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    RETURN QUERY
    WITH source_rows AS (
        SELECT
            subscription.id AS subscription_id,
            subscription.revision AS subscription_revision,
            pg_catalog.concat_ws(
                pg_catalog.chr(31),
                platform.billing_revenue_projection_id_v1(
                    subscription.id
                )::text,
                subscription.id::text,
                subscription.revision::text,
                subscription.offering_version_id::text,
                subscription.price_version_id::text,
                pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
                    pg_catalog.concat_ws(
                        pg_catalog.chr(31), price.id::text,
                        subscription.offering_version_id::text,
                        price.cadence, price.minor_units::text,
                        price.currency::text,
                        price.currency_registry_version::text,
                        price.money_policy_version
                    ), 'UTF8'
                )), 'hex'),
                subscription.state, price.cadence,
                price.currency::text,
                price.currency_registry_version::text,
                price.minor_units::text,
                platform.billing_revenue_monthly_v1(
                    price.minor_units, price.cadence
                )::text,
                CASE WHEN price.cadence = 'monthly'
                    THEN (price.minor_units * 12)::text
                    ELSE price.minor_units::text
                END
            ) AS canonical_row
        FROM platform.billing_subscriptions AS subscription
        JOIN platform.billing_price_versions AS price
          ON price.id = subscription.price_version_id
         AND price.offering_version_id = subscription.offering_version_id
        WHERE subscription.state IN ('active', 'past_due')
          AND price.cadence IN ('monthly', 'yearly')
    )
    SELECT
        pg_catalog.count(*)::bigint,
        COALESCE(
            pg_catalog.max(subscription_revision), 0::bigint
        ),
        pg_catalog.sha256(pg_catalog.convert_to(
            COALESCE(pg_catalog.string_agg(
                canonical_row, pg_catalog.chr(30)
                ORDER BY subscription_id
            ), ''), 'UTF8'
        ))
    FROM source_rows;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.billing_revenue_source_digest_v1(uuid) FROM PUBLIC;

CREATE FUNCTION platform.billing_revenue_projection_digest_v1(p_authority_id uuid)
RETURNS TABLE (
    row_count bigint,
    revision_ceiling bigint,
    canonical_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    RETURN QUERY
    SELECT
        pg_catalog.count(*)::bigint,
        COALESCE(
            pg_catalog.max(subscription_revision), 0::bigint
        ),
        pg_catalog.sha256(pg_catalog.convert_to(
            COALESCE(pg_catalog.string_agg(
                pg_catalog.concat_ws(
                    pg_catalog.chr(31), id::text,
                    subscription_id::text,
                    subscription_revision::text,
                    offering_version_id::text,
                    price_version_id::text,
                    pg_catalog.encode(price_source_sha256, 'hex'),
                    state, cadence, currency::text,
                    currency_registry_version::text,
                    source_minor_units::text,
                    monthly_minor_units::text,
                    annual_minor_units::text
                ),
                pg_catalog.chr(30) ORDER BY subscription_id
            ), ''), 'UTF8'
        ))
    FROM platform.billing_revenue_projection AS projection
    WHERE projection.authority_id = p_authority_id;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.billing_revenue_projection_digest_v1(uuid) FROM PUBLIC;
-- +goose StatementEnd
