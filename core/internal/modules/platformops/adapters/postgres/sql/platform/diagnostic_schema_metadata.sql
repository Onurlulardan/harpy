SELECT pg_catalog.count(*)::text,
       pg_catalog.count(*) FILTER (WHERE nspname = 'platform')::text,
       pg_catalog.count(*) FILTER (WHERE nspname LIKE 'tenant\_%' ESCAPE '\')::text,
       COALESCE(pg_catalog.max(nspname), '')
FROM pg_catalog.pg_namespace
WHERE nspname = 'platform' OR nspname LIKE 'tenant\_%' ESCAPE '\'
