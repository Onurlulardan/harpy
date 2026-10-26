INSERT INTO work_labels (
    id, board_id, name, color, row_version, deleted_at, created_at, updated_at
) VALUES ($1,$2,$3,$4,1,NULL,transaction_timestamp(),transaction_timestamp())
ON CONFLICT (id) DO UPDATE SET
    board_id=EXCLUDED.board_id, name=EXCLUDED.name, color=EXCLUDED.color,
    deleted_at=NULL, updated_at=EXCLUDED.updated_at
