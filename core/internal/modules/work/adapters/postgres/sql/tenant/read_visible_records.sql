WITH supplied_principals AS (
    SELECT principal.kind, principal.id
    FROM ROWS FROM (
        pg_catalog.unnest($3::text[]),
        pg_catalog.unnest($4::uuid[])
    ) AS principal(kind, id)
), visible_memberships AS MATERIALIZED (
    SELECT membership.id, membership.board_id, membership.principal_kind,
           membership.principal_id, membership.board_role,
           membership.row_version, membership.created_at
    FROM work_board_memberships AS membership
    JOIN supplied_principals AS principal
      ON principal.kind = membership.principal_kind
     AND principal.id = membership.principal_id
    WHERE membership.active
), visible_boards AS (
    SELECT board.*, pg_catalog.max(CASE membership.board_role
        WHEN 'VIEWER' THEN 1 WHEN 'EDITOR' THEN 2 WHEN 'ADMIN' THEN 3 ELSE 0 END) AS access_rank
    FROM work_boards AS board
    JOIN visible_memberships AS membership ON membership.board_id = board.id
    GROUP BY board.id
), records AS (
    SELECT board.id AS record_id, board.created_at AS sort_at,
        pg_catalog.jsonb_build_object(
            'id', board.id, 'name', board.name, 'description', board.description,
            'creator_id', board.creator_user_id, 'lifecycle', board.lifecycle_state,
            'row_version', board.row_version, 'created_at', board.created_at,
			'updated_at', board.updated_at,
            'task_count', (SELECT pg_catalog.count(*) FROM work_tasks AS task
                WHERE task.board_id = board.id AND task.deleted_at IS NULL),
            'member_count', (SELECT pg_catalog.count(*) FROM work_board_memberships AS member_row
                WHERE member_row.board_id = board.id AND member_row.active),
            'my_role', CASE board.access_rank WHEN 3 THEN 'ADMIN'
                WHEN 2 THEN 'EDITOR' ELSE 'VIEWER' END
		) || CASE WHEN $15 THEN pg_catalog.jsonb_build_object(
			'created_by', board.creator_user_id,
			'organization_id', (SELECT member_row.principal_id
				FROM work_board_memberships AS member_row
				WHERE member_row.board_id = board.id AND member_row.active
				  AND member_row.principal_kind = 'ORGANIZATION'
				ORDER BY member_row.principal_id LIMIT 1)
		) ELSE '{}'::jsonb END AS document
    FROM visible_boards AS board
    WHERE $1 = 'boards' AND ($5 = '00000000-0000-0000-0000-000000000000'::uuid OR board.id = $5)
    UNION ALL
    SELECT task.id, COALESCE(task.due_at, task.created_at), pg_catalog.jsonb_build_object(
        'id', task.id, 'board_id', task.board_id, 'column_id', task.column_id,
        'kind', task.task_kind, 'status', task.status_category, 'title', task.title,
        'description', task.description, 'position', task.position,
        'priority', task.priority, 'creator_id', task.creator_user_id,
        'row_version', task.row_version, 'starts_at', task.starts_at, 'due_at', task.due_at,
        'created_at', task.created_at, 'updated_at', task.updated_at
	) || CASE WHEN $15 THEN pg_catalog.jsonb_build_object(
		'task_type', task.task_kind,
		'status', CASE task.status_category WHEN 'OPEN' THEN 'TODO' ELSE task.status_category END,
		'column_order', task.position, 'created_by', task.creator_user_id,
		'assignee_id', (SELECT assignee.principal_id FROM work_task_assignees AS assignee
			WHERE assignee.task_id = task.id AND assignee.active
			ORDER BY CASE assignee.principal_kind WHEN 'USER' THEN 0 ELSE 1 END,
				assignee.principal_kind, assignee.principal_id LIMIT 1),
		'assignee_type', COALESCE((SELECT assignee.principal_kind FROM work_task_assignees AS assignee
			WHERE assignee.task_id = task.id AND assignee.active
			ORDER BY CASE assignee.principal_kind WHEN 'USER' THEN 0 ELSE 1 END,
				assignee.principal_kind, assignee.principal_id LIMIT 1), ''),
		'assignee_info', NULL, 'due_date', task.due_at, 'start_date', task.starts_at,
		'completed_at', task.completed_at,
		'form_filter', metadata.form_filter,
		'is_recurring', COALESCE(metadata.is_recurring, false),
		'recurrence_rule', COALESCE(metadata.recurrence_rule, ''),
		'recurrence_end', metadata.recurrence_end,
		'labels', COALESCE((SELECT pg_catalog.jsonb_agg(label.name ORDER BY label.name)
			FROM work_task_labels AS link JOIN work_labels AS label ON label.id = link.label_id
			WHERE link.task_id = task.id AND link.active AND label.deleted_at IS NULL), '[]'::jsonb),
		'checklist', COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
			'id', item.id, 'text', item.item_text, 'checked', item.completed)
			ORDER BY item.position, item.id) FROM work_checklist_items AS item
			WHERE item.task_id = task.id AND item.deleted_at IS NULL), '[]'::jsonb),
		'comment_count', (SELECT pg_catalog.count(*) FROM work_comments AS comment
			WHERE comment.task_id = task.id AND comment.state = 'VISIBLE'),
		'workflow_task_id', NULL
	) ELSE '{}'::jsonb END
	FROM work_tasks AS task JOIN visible_boards AS board ON board.id = task.board_id
	LEFT JOIN work_task_compatibility_metadata AS metadata ON metadata.task_id = task.id
    WHERE $1 IN ('tasks','inbox') AND task.deleted_at IS NULL
      AND ($5 = '00000000-0000-0000-0000-000000000000'::uuid OR task.id = $5)
      AND ($6 = '00000000-0000-0000-0000-000000000000'::uuid OR task.board_id = $6)
      AND ($8 = '' OR task.status_category = $8)
	  AND ($14 = '' OR task.priority = $14)
      AND ($9 = '00000000-0000-0000-0000-000000000000'::uuid OR task.board_id = $9)
      AND ($10::timestamptz IS NULL OR task.due_at >= $10)
      AND ($11::timestamptz IS NULL OR task.due_at <= $11)
      AND ($1 <> 'inbox' OR EXISTS (
          SELECT 1 FROM work_task_assignees AS assignee
          JOIN supplied_principals AS principal
            ON principal.kind = assignee.principal_kind AND principal.id = assignee.principal_id
          WHERE assignee.task_id = task.id AND assignee.active
      ))
    UNION ALL
    SELECT comment.id, comment.created_at, pg_catalog.jsonb_build_object(
        'id', comment.id, 'board_id', comment.board_id, 'task_id', comment.task_id,
        'author_id', comment.author_id, 'body', comment.body, 'state', comment.state,
        'redaction_kind', comment.redaction_kind, 'redaction_reason', comment.redaction_reason,
        'row_version', comment.row_version, 'created_at', comment.created_at,
		'redacted_at', comment.redacted_at
	) || CASE WHEN $15 THEN pg_catalog.jsonb_build_object(
		'content', comment.body, 'user_id', comment.author_id, 'user_info', NULL
	) ELSE '{}'::jsonb END
    FROM work_comments AS comment JOIN visible_boards AS board ON board.id = comment.board_id
    WHERE $1 = 'comments' AND comment.task_id = $6
      AND ($5 = '00000000-0000-0000-0000-000000000000'::uuid OR comment.id = $5)
    UNION ALL
    SELECT event.id, COALESCE(event.starts_at, event.start_date::timestamptz),
      pg_catalog.jsonb_build_object(
        'id', event.id, 'owner_id', event.owner_id, 'audience', event.audience,
        'board_id', event.board_id, 'organization_id', event.organization_id,
        'title', event.title, 'description', event.description,
		'color', event.color,
        'starts_at', event.starts_at, 'ends_at', event.ends_at,
        'start_date', event.start_date, 'end_date', event.end_date,
        'timezone', event.timezone_name, 'generated', event.generated,
        'row_version', event.row_version, 'created_at', event.created_at,
		'updated_at', event.updated_at
	) || CASE WHEN $15 THEN pg_catalog.jsonb_build_object(
		'created_by', event.owner_id,
		'start_date', COALESCE(event.starts_at, event.start_date::timestamptz),
		'end_date', COALESCE(event.ends_at, event.end_date::timestamptz),
		'all_day', event.start_date IS NOT NULL, 'color', event.color,
		'form_slug', '', 'form_data_id', NULL
	) ELSE '{}'::jsonb END
    FROM work_calendar_events AS event
    LEFT JOIN visible_boards AS board ON board.id = event.board_id
    WHERE $1 = 'calendar' AND event.deleted_at IS NULL
      AND ($5 = '00000000-0000-0000-0000-000000000000'::uuid OR event.id = $5)
      AND (NOT $12 OR event.owner_id = $2)
      AND ($10::timestamptz IS NULL OR (
          (event.starts_at IS NOT NULL AND event.ends_at > $10) OR
          (event.start_date IS NOT NULL AND event.end_date > $10::date)
      ))
      AND ($11::timestamptz IS NULL OR (
          (event.starts_at IS NOT NULL AND event.starts_at < $11) OR
          (event.start_date IS NOT NULL AND event.start_date < $11::date)
      ))
      AND (event.owner_id = $2 OR board.id IS NOT NULL OR EXISTS (
          SELECT 1 FROM supplied_principals AS principal
          WHERE principal.kind = 'ORGANIZATION' AND principal.id = event.organization_id
      ))
    UNION ALL
    SELECT membership.id, membership.created_at, pg_catalog.jsonb_build_object(
        'id', membership.id, 'board_id', membership.board_id,
        'principal_kind', membership.principal_kind,
        'principal_id', membership.principal_id, 'role', membership.board_role,
        'row_version', membership.row_version, 'created_at', membership.created_at
	) || CASE WHEN $15 THEN pg_catalog.jsonb_build_object(
		'user_id', CASE WHEN membership.principal_kind = 'USER'
			THEN membership.principal_id ELSE NULL END,
		'user', NULL
	) ELSE '{}'::jsonb END
    FROM work_board_memberships AS membership
    JOIN visible_boards AS board ON board.id = membership.board_id
    WHERE $1 = 'memberships' AND membership.active AND membership.board_id = $6
), limited AS (
    SELECT records.document FROM records
    ORDER BY CASE WHEN $13 THEN records.sort_at END ASC NULLS LAST,
        CASE WHEN NOT $13 THEN records.sort_at END DESC NULLS LAST,
        records.record_id DESC LIMIT $7
)
SELECT COALESCE(pg_catalog.jsonb_agg(limited.document), '[]'::jsonb) FROM limited;
