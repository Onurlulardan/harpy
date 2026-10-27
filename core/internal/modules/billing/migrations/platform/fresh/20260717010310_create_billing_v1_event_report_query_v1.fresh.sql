-- +goose Up
-- Migration version: 20260717010310 UTC
-- +goose StatementBegin
CREATE FUNCTION platform.billing_assert_retained_tenant_authority_v1(
    p_authority_id uuid, p_tenant_id uuid
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    expected_writer text;
    expected_state text;
BEGIN
    IF p_authority_id IS NULL OR p_tenant_id IS NULL
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing retained tenant authority input is invalid';
    END IF;
    IF SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime' THEN
        expected_writer := 'v2'; expected_state := 'v2';
    ELSIF SESSION_USER IS NOT DISTINCT FROM 'harpy_adopter' THEN
        expected_writer := 'v1'; expected_state := 'shadow';
    ELSE
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing retained tenant authority session is invalid';
    END IF;
    PERFORM 1
    FROM platform.billing_tenant_authorities AS authority_row
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = authority_row.scope_kind
     AND group_row.scope_id = authority_row.scope_id
     AND group_row.group_id = authority_row.authority_group_id
    JOIN platform.domain_cutovers AS billing_row
      ON billing_row.scope_kind = authority_row.scope_kind
     AND billing_row.scope_id = authority_row.scope_id
     AND billing_row.domain = authority_row.billing_domain
    JOIN platform.domain_cutovers AS forms_row
      ON forms_row.scope_kind = authority_row.scope_kind
     AND forms_row.scope_id = authority_row.scope_id
     AND forms_row.domain = authority_row.forms_domain
    WHERE authority_row.authority_id = p_authority_id
      AND authority_row.tenant_id = p_tenant_id
      AND authority_row.authority_group_generation = group_row.generation
      AND authority_row.billing_generation = billing_row.generation
      AND authority_row.forms_generation = forms_row.generation
      AND group_row.writer = expected_writer AND group_row.state = expected_state
      AND billing_row.writer = expected_writer AND billing_row.state = expected_state
      AND forms_row.writer = expected_writer AND forms_row.state = expected_state
    FOR SHARE OF authority_row, group_row, billing_row, forms_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing retained tenant authority is not current';
    END IF;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_assert_retained_tenant_authority_v1(
    uuid, uuid
) FROM PUBLIC;
-- +goose StatementEnd

CREATE VIEW platform.billing_v1_event_report_query_v1
WITH (security_barrier = true)
AS
WITH current_global AS (
    SELECT authority.authority_id
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
      AND group_row.writer = 'v2' AND group_row.state = 'v2'
      AND domain_row.writer = 'v2' AND domain_row.state = 'v2'
), current_tenants AS (
    SELECT authority.authority_id, authority.tenant_id
    FROM platform.billing_tenant_authorities AS authority
    JOIN platform.tenants AS tenant ON tenant.id = authority.tenant_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = authority.scope_kind
     AND group_row.scope_id = authority.scope_id
     AND group_row.group_id = authority.authority_group_id
     AND group_row.generation = authority.authority_group_generation
    JOIN platform.domain_cutovers AS billing_row
      ON billing_row.scope_kind = authority.scope_kind
     AND billing_row.scope_id = authority.scope_id
     AND billing_row.domain = authority.billing_domain
     AND billing_row.generation = authority.billing_generation
    JOIN platform.domain_cutovers AS forms_row
      ON forms_row.scope_kind = authority.scope_kind
     AND forms_row.scope_id = authority.scope_id
     AND forms_row.domain = authority.forms_domain
     AND forms_row.generation = authority.forms_generation
    WHERE group_row.writer = 'v2' AND group_row.state = 'v2'
      AND billing_row.writer = 'v2' AND billing_row.state = 'v2'
      AND forms_row.writer = 'v2' AND forms_row.state = 'v2'
)
SELECT archive.source_event_id AS id, archive.tenant_id,
       archive.source_subscription_id AS subscription_id,
       archive.event_type, archive.amount_cents, archive.currency,
       archive.description, archive.metadata, archive.created_at
FROM current_tenants AS authority
JOIN platform.billing_v1_event_archive AS archive
  ON archive.authority_id = authority.authority_id
 AND archive.tenant_id = authority.tenant_id
UNION ALL
SELECT audit.id, audit.tenant_id,
       COALESCE(alias.legacy_id, audit.subscription_id) AS subscription_id,
       CASE
         WHEN audit.kind = 'cancellation_scheduled'
              OR audit.reason = 'cancellation' THEN 'subscription_cancelled'
         WHEN audit.kind = 'cancellation_resumed' THEN 'subscription_resumed'
         WHEN audit.kind = 'trial_claimed' THEN 'trial_started'
         WHEN audit.kind = 'manual_change_scheduled'
              OR audit.reason = 'manual_change' THEN 'plan_changed'
         WHEN audit.reason = 'trial_end' THEN 'trial_ended'
         WHEN subscription.state = 'expired' THEN 'subscription_expired'
         WHEN subscription.state = 'cancelled' THEN 'subscription_cancelled'
         ELSE 'subscription_renewed'
       END AS event_type,
       0::bigint AS amount_cents,
       pg_catalog.btrim(price.currency::text) AS currency,
       audit.reason AS description,
       pg_catalog.jsonb_build_object(
           'source_kind', audit.source_kind,
           'before_revision', audit.before_revision,
           'after_revision', audit.after_revision,
           'source_operation_id', audit.source_operation_id
       ) AS metadata,
       audit.occurred_at AS created_at
FROM current_global AS global_authority
JOIN current_tenants AS authority ON true
JOIN platform.billing_subscription_audit AS audit
  ON audit.authority_id = authority.authority_id
 AND audit.tenant_id = authority.tenant_id
JOIN platform.billing_subscriptions AS subscription
  ON subscription.id = audit.subscription_id
 AND subscription.tenant_id = audit.tenant_id
JOIN platform.billing_price_versions AS price
  ON price.authority_id = global_authority.authority_id
 AND price.id = subscription.price_version_id
LEFT JOIN platform.billing_legacy_identity_alias_v1 AS alias
  ON alias.authority_id = authority.authority_id
 AND alias.scope_kind = 'tenant'
 AND alias.scope_id = audit.tenant_id
 AND alias.record_kind = 'subscription'
 AND alias.canonical_id = audit.subscription_id
WHERE SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_runtime';

GRANT SELECT ON platform.billing_v1_event_report_query_v1
    TO harpy_runtime;
