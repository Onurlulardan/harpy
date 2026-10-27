SELECT change_row.id, change_row.decision_id, change_row.subscription_id,
       change_row.before_offering_version_id, change_row.before_price_version_id,
       change_row.after_offering_version_id, change_row.after_price_version_id,
       change_row.source_operation_id, change_row.state,
       change_row.effective_at, change_row.created_at, change_row.revision
FROM platform.billing_subscription_changes AS change_row
WHERE change_row.authority_id = $1 AND change_row.tenant_id = $2
  AND change_row.subscription_id = $3 AND change_row.state = 'scheduled'
ORDER BY change_row.effective_at, change_row.id
LIMIT 1
FOR UPDATE OF change_row;
