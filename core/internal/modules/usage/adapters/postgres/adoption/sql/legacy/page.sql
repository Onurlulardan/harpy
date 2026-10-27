SELECT logs.id,COALESCE(logs.tenant_id,$1),logs.user_id,
logs.method,logs.path,logs.status_code,logs.duration_ms,logs.start_time
FROM trace_logs AS logs
WHERE logs.user_id IS NOT NULL AND logs.method IN ('GET','HEAD','POST','PUT','PATCH','DELETE','OPTIONS')
AND logs.path IS NOT NULL AND logs.path <> '' AND logs.status_code BETWEEN 100 AND 599
AND logs.duration_ms BETWEEN 0 AND 86400000 AND logs.start_time IS NOT NULL
AND ($2::uuid IS NULL OR logs.id > $2) ORDER BY logs.id LIMIT $3
