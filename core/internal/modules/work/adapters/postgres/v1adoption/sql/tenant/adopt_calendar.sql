INSERT INTO work_calendar_events (
    id, owner_id, audience, board_id, organization_id, title, description, color,
    starts_at, ends_at, start_date, end_date, timezone_name, generated,
    edited, live_reference_count, legal_hold, row_version, deleted_at,
    created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,false,false,0,false,1,$14,$15,$16)
ON CONFLICT (id) DO UPDATE SET
    owner_id=EXCLUDED.owner_id, audience=EXCLUDED.audience,
    board_id=EXCLUDED.board_id, organization_id=EXCLUDED.organization_id,
    title=EXCLUDED.title, description=EXCLUDED.description, color=EXCLUDED.color,
    starts_at=EXCLUDED.starts_at, ends_at=EXCLUDED.ends_at,
    start_date=EXCLUDED.start_date, end_date=EXCLUDED.end_date,
    timezone_name=EXCLUDED.timezone_name, deleted_at=EXCLUDED.deleted_at,
    created_at=EXCLUDED.created_at, updated_at=EXCLUDED.updated_at
