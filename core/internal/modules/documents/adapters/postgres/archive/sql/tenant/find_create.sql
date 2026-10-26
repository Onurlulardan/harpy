SELECT a.id
FROM documents_operations o
JOIN documents_archives a ON a.id = o.aggregate_id
WHERE o.actor_id = $1::uuid
  AND o.operation_kind = 'documents.archive.create'
  AND o.idempotency_key = $2 AND o.request_sha256 = $3
  AND o.aggregate_kind = 'ARCHIVE';
