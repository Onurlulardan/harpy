INSERT INTO platform.billing_entitlement_decisions (
    id, authority_id, tenant_id, actor_id, subscription_id,
    entitlement_definition_id, entitlement_key, meter_period_id,
    source_operation_id, requested_amount, current_amount, limit_amount,
    unlimited, allowed, soft_exceeded, reason, correlation_id,
    decided_at, retain_until
)
SELECT $4, authority.authority_id, authority.tenant_id, $3,
       subscription.id, definition.id, definition.entitlement_key,
       period.id, $7::uuid, $8, $9, $10, $11, $12, $13,
       CASE
         WHEN $13 THEN 'soft_limit'
         WHEN $12 THEN 'allowed'
         ELSE 'hard_limit'
       END,
       $7::uuid::text, $14::timestamptz,
       $14::timestamptz + interval '7 years'
FROM platform.billing_tenant_authorities AS authority
JOIN platform.billing_meter_periods AS period
  ON period.authority_id = authority.authority_id
 AND period.tenant_id = authority.tenant_id
JOIN platform.billing_entitlement_definitions AS definition
  ON definition.id = period.entitlement_definition_id
 AND definition.entitlement_key = period.entitlement_key
LEFT JOIN LATERAL (
    SELECT candidate.id
    FROM platform.billing_subscriptions AS candidate
    WHERE candidate.authority_id = authority.authority_id
      AND candidate.tenant_id = authority.tenant_id
      AND candidate.state IN ('trialing', 'active', 'past_due')
    LIMIT 1
) AS subscription ON true
WHERE authority.authority_id = $1 AND authority.tenant_id = $2
  AND definition.entitlement_key = $5 AND period.id = $6
RETURNING id;
