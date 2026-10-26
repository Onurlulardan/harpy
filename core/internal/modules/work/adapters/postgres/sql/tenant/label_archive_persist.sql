UPDATE work_labels
SET deleted_at = $3, row_version = row_version + 1,
    updated_at = transaction_timestamp()
WHERE id = $1 AND row_version = $2
  AND ((deleted_at IS NULL) <> ($3::timestamptz IS NULL))
