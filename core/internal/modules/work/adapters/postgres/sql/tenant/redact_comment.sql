UPDATE work_comments
SET body = '', state = 'REDACTED', redaction_kind = $2, redacted_by = $3,
    redaction_reason = NULLIF($4, ''), row_version = $5, redacted_at = $6
WHERE id = $1 AND row_version = $7 AND state = 'VISIBLE'
