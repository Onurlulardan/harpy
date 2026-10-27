SELECT offering_version.id, price.id, price.currency::text,
       offering_version.trial_days, price.cadence
FROM platform.billing_offering_versions AS offering_version
JOIN platform.billing_price_versions AS price
  ON price.offering_version_id = offering_version.id
WHERE offering_version.id = $1 AND price.id = $2;
