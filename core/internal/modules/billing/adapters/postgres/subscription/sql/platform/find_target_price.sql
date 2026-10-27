SELECT offering_version.id, price.id, price.currency::text,
       offering_version.trial_days, price.cadence
FROM platform.billing_offerings AS offering
JOIN platform.billing_offering_versions AS offering_version
  ON offering_version.offering_id = offering.id
JOIN platform.billing_price_versions AS price
  ON price.offering_version_id = offering_version.id
WHERE offering_version.id = $1 AND price.id = $2
  AND offering.lifecycle = 'active'
  AND offering_version.valid_start <= pg_catalog.statement_timestamp()
  AND (offering_version.valid_end IS NULL
       OR offering_version.valid_end > pg_catalog.statement_timestamp())
  AND price.effective_start <= pg_catalog.statement_timestamp()
  AND (price.effective_end IS NULL
       OR price.effective_end > pg_catalog.statement_timestamp());
