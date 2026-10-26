INSERT INTO platform.platform_migration_history (
    version,
    record_kind,
    is_applied,
    applied_at
) VALUES (
    0,
    'anchor',
    TRUE,
    pg_catalog.statement_timestamp()
)
ON CONFLICT (version) DO NOTHING;
