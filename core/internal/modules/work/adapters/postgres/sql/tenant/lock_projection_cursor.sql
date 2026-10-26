SELECT through_sequence
FROM work_projection_streams
WHERE source_owner = $1 AND aggregate_kind = $2 AND aggregate_id = $3
FOR UPDATE
