INSERT INTO work_comments (
    id, board_id, task_id, author_id, body, state, redaction_kind,
    redacted_by, redaction_reason, row_version, created_at, redacted_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,1,$10,$11)
