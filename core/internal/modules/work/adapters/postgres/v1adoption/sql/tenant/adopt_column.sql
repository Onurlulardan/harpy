INSERT INTO work_board_columns (
    id, board_id, name, color, status_category, position, row_version, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,1,$7,$8)
ON CONFLICT (id) DO UPDATE SET
    board_id=EXCLUDED.board_id, name=EXCLUDED.name, color=EXCLUDED.color,
    status_category=EXCLUDED.status_category, position=EXCLUDED.position,
    created_at=EXCLUDED.created_at, updated_at=EXCLUDED.updated_at
