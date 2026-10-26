UPDATE platform.billing_subscription_changes AS change_row
SET state = 'applied', resolved_at = $6, revision = change_row.revision + 1
WHERE change_row.authority_id = $1 AND change_row.tenant_id = $2
  AND change_row.id = $3 AND change_row.subscription_id = $4
  AND change_row.revision = $5 AND change_row.state = 'scheduled'
RETURNING change_row.id;
