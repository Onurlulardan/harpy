SELECT label_id
FROM work_task_labels
WHERE task_id = $1 AND active
ORDER BY label_id
FOR UPDATE
