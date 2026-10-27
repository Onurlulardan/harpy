SELECT pg_catalog.pg_advisory_xact_lock(
    pg_catalog.hashtextextended(
        'billing-discount-code|' || discount_row.normalized_code, 0
    )
)
FROM platform.billing_discounts AS discount_row
WHERE discount_row.normalized_code = $1::text
  AND discount_row.lifecycle = 'active';
