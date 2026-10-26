SELECT id, board_id, task_id, author_id, body, state, redaction_kind,
       redacted_by, redaction_reason, row_version, created_at, redacted_at
FROM work_comments
WHERE id = $1
FOR UPDATE
