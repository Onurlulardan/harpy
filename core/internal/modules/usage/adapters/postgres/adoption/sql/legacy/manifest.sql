SELECT relation_oid::bigint,source_count,source_root FROM (
SELECT pg_catalog.to_regclass($1)::oid AS relation_oid,
pg_catalog.count(*) AS source_count,
pg_catalog.sha256(pg_catalog.convert_to(COALESCE(pg_catalog.string_agg(
pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(pg_catalog.jsonb_build_array(
logs.id,COALESCE(logs.tenant_id,$2),logs.user_id,logs.method,logs.path,
logs.status_code,logs.duration_ms,logs.start_time)::text,'UTF8')),'hex'),'' ORDER BY logs.id),
'zero'),'UTF8')) AS source_root FROM trace_logs AS logs
WHERE logs.user_id IS NOT NULL AND logs.method IN ('GET','HEAD','POST','PUT','PATCH','DELETE','OPTIONS')
AND logs.path IS NOT NULL AND logs.path <> '' AND logs.status_code BETWEEN 100 AND 599
AND logs.duration_ms BETWEEN 0 AND 86400000 AND logs.start_time IS NOT NULL) AS manifest
