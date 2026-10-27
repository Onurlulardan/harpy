UPDATE notifications_deliveries
SET state = $3, attempt = $4, lease_owner = $5, lease_generation = $6,
    lease_until = $7, not_before = $8, outcome_code = $9,
    outcome_unknown = $10, row_version = $11
WHERE id = $1::uuid AND row_version = $2;
