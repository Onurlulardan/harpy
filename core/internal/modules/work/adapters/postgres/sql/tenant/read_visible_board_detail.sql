WITH supplied AS (
    SELECT principal.kind, principal.id
    FROM ROWS FROM (
        pg_catalog.unnest($2::text[]),
        pg_catalog.unnest($3::uuid[])
    ) AS principal(kind, id)
), selected_board AS MATERIALIZED (
    SELECT board.*
    FROM work_boards AS board
    WHERE board.id = $1 AND EXISTS (
        SELECT 1 FROM work_board_memberships AS membership
        JOIN supplied AS principal
          ON principal.kind = membership.principal_kind
         AND principal.id = membership.principal_id
        WHERE membership.board_id = board.id AND membership.active
    )
), task_rows AS (
    SELECT task.*, assignee.principal_id AS assignee_id,
           assignee.principal_kind AS assignee_kind,
           metadata.form_filter AS compatibility_form_filter,
           metadata.is_recurring AS compatibility_is_recurring,
           metadata.recurrence_rule AS compatibility_recurrence_rule,
           metadata.recurrence_end AS compatibility_recurrence_end,
           pg_catalog.row_number() OVER (
               PARTITION BY task.column_id ORDER BY task.position, task.id
           ) - 1 AS legacy_order
    FROM work_tasks AS task
    JOIN selected_board AS board ON board.id = task.board_id
    LEFT JOIN work_task_compatibility_metadata AS metadata ON metadata.task_id = task.id
    LEFT JOIN LATERAL (
        SELECT item.principal_id, item.principal_kind
        FROM work_task_assignees AS item
        WHERE item.task_id = task.id AND item.active
        ORDER BY CASE item.principal_kind WHEN 'USER' THEN 0 ELSE 1 END,
                 item.principal_kind, item.principal_id
        LIMIT 1
    ) AS assignee ON TRUE
    WHERE task.deleted_at IS NULL
), task_documents AS (
    SELECT task.column_id, task.position, task.id,
      pg_catalog.jsonb_build_object(
        'id', task.id, 'board_id', task.board_id, 'column_id', task.column_id,
        'column_order', task.legacy_order, 'title', task.title,
        'description', task.description, 'task_type', task.task_kind,
        'priority', task.priority,
        'status', CASE task.status_category WHEN 'OPEN' THEN 'TODO' ELSE task.status_category END,
		'form_filter', task.compatibility_form_filter,
        'created_by', task.creator_user_id, 'assignee_id', task.assignee_id,
        'assignee_type', COALESCE(task.assignee_kind, ''), 'assignee_info', NULL,
        'due_date', task.due_at, 'start_date', task.starts_at,
        'completed_at', task.completed_at, 'created_at', task.created_at,
		'updated_at', task.updated_at,
		'is_recurring', COALESCE(task.compatibility_is_recurring, false),
		'recurrence_rule', COALESCE(task.compatibility_recurrence_rule, ''),
		'recurrence_end', task.compatibility_recurrence_end,
        'labels', COALESCE((SELECT pg_catalog.jsonb_agg(label.name ORDER BY label.name)
            FROM work_task_labels AS link JOIN work_labels AS label ON label.id = link.label_id
            WHERE link.task_id = task.id AND link.active AND label.deleted_at IS NULL), '[]'::jsonb),
        'checklist', COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
                'id', item.id, 'text', item.item_text, 'checked', item.completed)
                ORDER BY item.position, item.id)
            FROM work_checklist_items AS item
            WHERE item.task_id = task.id AND item.deleted_at IS NULL), '[]'::jsonb),
        'comment_count', (SELECT pg_catalog.count(*) FROM work_comments AS comment
            WHERE comment.task_id = task.id AND comment.state = 'VISIBLE'),
        'workflow_task_id', NULL
      ) AS document
    FROM task_rows AS task
), column_documents AS (
    SELECT board_column.id, board_column.position,
      pg_catalog.jsonb_build_object(
        'id', board_column.id, 'board_id', board_column.board_id, 'name', board_column.name,
        'order', pg_catalog.row_number() OVER (
            ORDER BY board_column.position, board_column.id
        ) - 1,
		'color', board_column.color,
        'linked_status', CASE board_column.status_category WHEN 'OPEN' THEN 'TODO'
            ELSE board_column.status_category END
      ) AS document,
      COALESCE((SELECT pg_catalog.jsonb_agg(task.document ORDER BY task.position, task.id)
          FROM task_documents AS task
          WHERE task.column_id = board_column.id), '[]'::jsonb) AS tasks
    FROM work_board_columns AS board_column
    JOIN selected_board AS board ON board.id = board_column.board_id
)
SELECT COALESCE((SELECT pg_catalog.jsonb_build_array(pg_catalog.jsonb_build_object(
    'board', pg_catalog.jsonb_build_object(
        'id', board.id, 'name', board.name, 'description', board.description,
        'organization_id', (SELECT membership.principal_id
            FROM work_board_memberships AS membership
            WHERE membership.board_id = board.id AND membership.active
              AND membership.principal_kind = 'ORGANIZATION'
            ORDER BY membership.principal_id LIMIT 1),
        'created_by', board.creator_user_id, 'created_at', board.created_at,
        'updated_at', board.updated_at,
        'task_count', (SELECT pg_catalog.count(*) FROM task_rows),
        'member_count', (SELECT pg_catalog.count(*) FROM work_board_memberships AS membership
            WHERE membership.board_id = board.id AND membership.active)
    ),
    'columns', COALESCE((SELECT pg_catalog.jsonb_agg(item.document ORDER BY item.position, item.id)
        FROM column_documents AS item), '[]'::jsonb),
    'tasks', COALESCE((SELECT pg_catalog.jsonb_object_agg(item.id::text, item.tasks)
        FROM column_documents AS item), '{}'::jsonb)
)) FROM selected_board AS board), '[]'::jsonb);
