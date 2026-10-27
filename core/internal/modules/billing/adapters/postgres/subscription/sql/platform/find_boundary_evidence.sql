SELECT evidence.outcome,
       CASE WHEN evidence.outcome = 'paid' THEN receipt.id END,
       evidence.next_period_start, evidence.next_period_end,
       evidence.grace_ends_at
FROM platform.billing_subscription_boundary_evidence AS evidence
JOIN platform.billing_subscription_boundary_jobs AS intent
  ON intent.id = evidence.boundary_job_id
 AND intent.tenant_id = evidence.tenant_id
JOIN platform.billing_provider_events AS event_row
  ON event_row.id = evidence.provider_event_id
 AND event_row.tenant_id = evidence.tenant_id
JOIN platform.billing_webhook_receipts AS receipt
  ON receipt.id = event_row.webhook_receipt_id
 AND receipt.tenant_id = event_row.tenant_id
JOIN platform.billing_provider_objects AS mapping
  ON mapping.id = event_row.provider_object_mapping_id
 AND mapping.tenant_id = event_row.tenant_id
WHERE evidence.authority_id = $1 AND evidence.tenant_id = $2
  AND evidence.boundary_job_id = $3 AND evidence.subscription_id = $4
  AND evidence.boundary_kind = $5 AND evidence.due_at = $6
  AND evidence.expected_revision = $7
  AND intent.subscription_id = evidence.subscription_id
  AND intent.boundary_kind = evidence.boundary_kind
  AND intent.run_at = evidence.due_at
  AND intent.expected_revision = evidence.expected_revision
  AND mapping.subscription_id = evidence.subscription_id
  AND event_row.ordering_mode = 'contiguous'
  AND receipt.processor_state = 'applied'
ORDER BY event_row.ordering_sequence DESC,
         event_row.normalized_at DESC, evidence.id DESC
LIMIT 1
FOR SHARE OF evidence, intent, event_row, receipt, mapping;
