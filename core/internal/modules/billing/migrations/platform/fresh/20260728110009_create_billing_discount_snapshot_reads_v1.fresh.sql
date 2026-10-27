-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_discount_eligibility_snapshot_read_v1(
    p_tenant_authority_id uuid,
    p_tenant_id uuid,
    p_code text,
    p_offering_version_id uuid,
    p_at timestamptz,
    p_global_authority_id uuid
)
RETURNS SETOF record
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
RETURN QUERY
WITH runtime_guard AS MATERIALIZED (
    SELECT 1
    WHERE SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
      AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_owner'
      AND p_tenant_authority_id IS NOT DISTINCT FROM
          platform.billing_resolve_current_tenant_authority_v1(p_tenant_id)
      AND p_global_authority_id IS NOT DISTINCT FROM
          platform.billing_resolve_current_global_authority_v1('billing'::text)
), current_candidates AS MATERIALIZED (
    SELECT subscription.*, pg_catalog.count(*) OVER () AS candidate_count
    FROM runtime_guard
    CROSS JOIN platform.billing_subscriptions AS subscription
    WHERE subscription.authority_id = p_tenant_authority_id
      AND subscription.tenant_id = p_tenant_id
      AND subscription.state IN ('trialing', 'active', 'past_due')
), current_subscription AS MATERIALIZED (
    SELECT * FROM current_candidates ORDER BY updated_at DESC, id LIMIT 1
), candidate AS MATERIALIZED (
    SELECT current_subscription.candidate_count,
           discount_row.id AS discount_id, discount_row.revision,
           version_row.id AS discount_version_id, version_row.version,
           current_subscription.id AS subscription_id,
           current_subscription.offering_version_id,
           current_subscription.price_version_id,
           price.currency::text AS subscription_currency,
           discount_row.normalized_code, version_row.discount_kind,
           version_row.fixed_minor_units, version_row.currency::text,
           version_row.currency_registry_version, version_row.basis_points,
           version_row.duration, version_row.repeating_cycles,
           version_row.priority, version_row.global_use_limit,
           version_row.per_tenant_use_limit, version_row.all_offerings,
           version_row.expected_offering_count,
           version_row.valid_start, version_row.valid_end,
           version_row.available
    FROM current_subscription
    JOIN platform.billing_price_versions AS price
      ON price.id = current_subscription.price_version_id
     AND price.offering_version_id = current_subscription.offering_version_id
    JOIN platform.billing_discounts AS discount_row
      ON discount_row.authority_id = p_global_authority_id
     AND discount_row.normalized_code = p_code
     AND discount_row.lifecycle = 'active'
    JOIN LATERAL (
        SELECT version_row.*
        FROM platform.billing_discount_versions AS version_row
        WHERE version_row.discount_id = discount_row.id
          AND version_row.valid_start <= p_at
          AND (version_row.valid_end IS NULL OR p_at < version_row.valid_end)
        ORDER BY version_row.version DESC, version_row.id
        LIMIT 1
    ) AS version_row ON true
    WHERE p_offering_version_id IS NULL
       OR current_subscription.offering_version_id = p_offering_version_id
)
SELECT p_tenant_id, candidate.candidate_count,
       candidate.discount_id, candidate.revision,
       candidate.discount_version_id, candidate.version,
       candidate.subscription_id, candidate.offering_version_id,
       candidate.price_version_id, candidate.subscription_currency,
       candidate.normalized_code, candidate.discount_kind,
       candidate.fixed_minor_units, candidate.currency,
       candidate.currency_registry_version, candidate.basis_points,
       candidate.duration, candidate.repeating_cycles, candidate.priority,
       candidate.global_use_limit, candidate.per_tenant_use_limit,
       candidate.all_offerings, candidate.expected_offering_count,
       candidate.valid_start, candidate.valid_end, candidate.available,
       (SELECT pg_catalog.count(*)
        FROM platform.billing_discount_redemptions AS redemption
        WHERE redemption.discount_version_id = candidate.discount_version_id),
       (SELECT pg_catalog.count(*)
        FROM platform.billing_discount_redemptions AS redemption
        WHERE redemption.discount_version_id = candidate.discount_version_id
          AND redemption.tenant_id = p_tenant_id),
       (candidate.all_offerings OR EXISTS (
          SELECT 1 FROM platform.billing_discount_offerings AS allowed
          WHERE allowed.discount_version_id = candidate.discount_version_id
            AND allowed.offering_version_id = candidate.offering_version_id
       )),
       EXISTS (
          SELECT 1 FROM platform.billing_discount_redemptions AS redemption
          WHERE redemption.subscription_id = candidate.subscription_id
            AND redemption.discount_version_id = candidate.discount_version_id
       )
FROM candidate;
END;
$function$;

CREATE FUNCTION platform.billing_discount_offerings_snapshot_read_v1(
    p_authority_id uuid, p_discount_version_id uuid
)
RETURNS SETOF uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
RETURN QUERY
SELECT offering_version_id
FROM platform.billing_discount_offerings
WHERE authority_id = p_authority_id
  AND discount_version_id = p_discount_version_id
  AND SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_owner'
  AND p_authority_id IS NOT DISTINCT FROM
      platform.billing_resolve_current_global_authority_v1('billing'::text)
ORDER BY offering_version_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.billing_discount_eligibility_snapshot_read_v1(
        uuid,uuid,text,uuid,timestamptz,uuid
    ) FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.billing_discount_offerings_snapshot_read_v1(uuid,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.billing_discount_eligibility_snapshot_read_v1(
        uuid,uuid,text,uuid,timestamptz,uuid
    ) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION
    platform.billing_discount_offerings_snapshot_read_v1(uuid,uuid)
    TO harpy_runtime;
