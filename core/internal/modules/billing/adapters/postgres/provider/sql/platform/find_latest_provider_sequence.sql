SELECT event_row.ordering_sequence
FROM platform.billing_provider_events AS event_row
WHERE event_row.authority_id = $1
  AND event_row.tenant_id = $2
  AND event_row.provider_object_mapping_id = $3
  AND event_row.ordering_mode = 'contiguous'
ORDER BY event_row.ordering_sequence DESC, event_row.id DESC
LIMIT 1
FOR SHARE OF event_row;
