SELECT target_kind, target_id, source_sha256, target_sha256
FROM work_adoption_objects
WHERE run_id = $1 AND state IN ('COPIED','MATCHED')
ORDER BY target_kind, target_id
