WITH candidates AS MATERIALIZED (
    SELECT period.id
    FROM platform.billing_meter_periods AS period
    JOIN platform.billing_entitlement_definitions AS definition
      ON definition.id = period.entitlement_definition_id
     AND definition.entitlement_key = period.entitlement_key
    WHERE period.authority_id = $1 AND period.tenant_id = $2
      AND period.entitlement_key = $3 AND period.state = 'open'
      AND period.period_start <= pg_catalog.statement_timestamp()
      AND period.period_end > pg_catalog.statement_timestamp()
      AND definition.value_type = 'integer'
      AND period.meter_kind IN ('counter', 'gauge')
    FOR SHARE OF period
)
SELECT (pg_catalog.array_agg(id ORDER BY id))[1]
FROM candidates
HAVING pg_catalog.count(*) = 1;
