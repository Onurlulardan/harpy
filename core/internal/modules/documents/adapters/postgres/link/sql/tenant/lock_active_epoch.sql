SELECT reference_epoch
FROM documents_items
WHERE id = $1::uuid AND lifecycle_state = 'ACTIVE'
FOR UPDATE;
