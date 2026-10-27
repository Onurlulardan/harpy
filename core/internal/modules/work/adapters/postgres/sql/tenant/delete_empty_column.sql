DELETE FROM work_board_columns AS column_row
WHERE column_row.id = $1 AND column_row.row_version = $2
  AND NOT EXISTS (
      SELECT 1 FROM work_tasks AS task WHERE task.column_id = column_row.id
  )
