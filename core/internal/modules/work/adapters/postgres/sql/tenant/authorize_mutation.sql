WITH supplied_principals AS (
    SELECT principal.kind, principal.id
    FROM ROWS FROM (
        pg_catalog.unnest($4::text[]),
        pg_catalog.unnest($5::uuid[])
    ) AS principal(kind, id)
), resource_context AS (
    SELECT board.id AS board_id, NULL::uuid AS owner_id,
           NULL::uuid AS organization_id, NULL::uuid AS author_id
    FROM work_boards AS board
    WHERE board.id = $2 AND (
        $1 LIKE 'work.board.%' OR $1 = 'work.membership.change' OR
        $1 IN ('work.column.create', 'work.column.update') OR $1 = 'work.task.create' OR
        $1 = 'work.label.change'
    )
    UNION ALL
    SELECT column_row.board_id, NULL, NULL, NULL
    FROM work_board_columns AS column_row
    WHERE column_row.id = $2 AND $1 = 'work.column.update'
    UNION ALL
    SELECT task.board_id, NULL, NULL, NULL
    FROM work_tasks AS task
    WHERE task.id = $2 AND (
        $1 LIKE 'work.task.%' OR $1 = 'work.comment.create' OR
        $1 = 'work.checklist.change' OR $1 = 'work.label.change'
    )
    UNION ALL
    SELECT comment.board_id, NULL, NULL, comment.author_id
    FROM work_comments AS comment
    WHERE comment.id = $2 AND $1 = 'work.comment.redact'
    UNION ALL
    SELECT event.board_id, event.owner_id, event.organization_id, NULL
    FROM work_calendar_events AS event
    WHERE event.id = $2 AND (
        $1 = 'work.calendar.change' OR $1 LIKE 'work.recurrence.%'
    )
    UNION ALL
    SELECT board.id, NULL, NULL, NULL
    FROM work_boards AS board
    WHERE board.id = $2 AND $1 = 'work.calendar.change'
    UNION ALL
    SELECT NULL, NULL, $2, NULL
    WHERE $1 = 'work.calendar.change'
    UNION ALL
    SELECT NULL, $3, NULL, NULL
    WHERE $2 = $3 AND $1 = 'work.calendar.change'
), matching_memberships AS MATERIALIZED (
    SELECT membership.board_id, membership.board_role
    FROM work_board_memberships AS membership
    JOIN supplied_principals AS principal
      ON principal.kind = membership.principal_kind
     AND principal.id = membership.principal_id
    WHERE membership.active
    FOR SHARE OF membership
), ranked_context AS (
    SELECT context_row.*,
        COALESCE(pg_catalog.max(CASE membership.board_role
            WHEN 'VIEWER' THEN 1 WHEN 'EDITOR' THEN 2 WHEN 'ADMIN' THEN 3
            ELSE 0 END), 0) AS board_rank,
        EXISTS (
            SELECT 1 FROM supplied_principals AS principal
            WHERE principal.kind = 'ORGANIZATION'
              AND principal.id = context_row.organization_id
        ) AS organization_member
    FROM resource_context AS context_row
    LEFT JOIN matching_memberships AS membership
      ON membership.board_id = context_row.board_id
    GROUP BY context_row.board_id, context_row.owner_id,
             context_row.organization_id, context_row.author_id
), required AS (
    SELECT CASE
        WHEN $1 = 'work.comment.create' THEN 1
        WHEN $1 IN ('work.board.update', 'work.board.archive', 'work.board.restore',
                    'work.membership.change', 'work.column.create',
                    'work.column.update') THEN 3
        ELSE 2
    END AS board_rank
)
SELECT CASE
    WHEN $1 IN ('work.projection.apply','work.projection.repair')
         AND $2 <> '00000000-0000-0000-0000-000000000000'::uuid
        THEN true
    WHEN $1 = 'work.board.create' AND $2 = '00000000-0000-0000-0000-000000000000'::uuid
        THEN true
    ELSE EXISTS (
        SELECT 1 FROM ranked_context AS context_row CROSS JOIN required
        WHERE (context_row.board_id IS NOT NULL AND
               context_row.board_rank >= required.board_rank) OR
              (context_row.owner_id = $3) OR
              context_row.organization_member OR
              ($1 = 'work.comment.redact' AND context_row.author_id = $3)
    )
END;
