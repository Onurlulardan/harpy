CREATE TABLE platform.platform_migration_history (
    version bigint PRIMARY KEY,
    record_kind text NOT NULL CHECK (record_kind IN ('anchor', 'migration')),
    is_applied boolean NOT NULL CHECK (is_applied),
    owner text,
    scope text,
    origin text,
    selected_path text,
    selected_sha256 text,
    run_id uuid,
    target_database_generation bigint,
    next_minimum_runtime_generation bigint,
    applied_at timestamptz NOT NULL,
    CONSTRAINT platform_migration_history_version_check CHECK (
        version = 0 OR version BETWEEN 20000101000000 AND 99991231235959
    ),
    CONSTRAINT platform_migration_history_owner_check CHECK (
        owner IS NULL OR owner ~ '^[a-z][a-z0-9_-]{0,63}$'
    ),
    CONSTRAINT platform_migration_history_scope_check CHECK (
        scope IS NULL OR scope IN ('platform', 'tenant')
    ),
    CONSTRAINT platform_migration_history_origin_check CHECK (
        origin IS NULL OR origin IN ('fresh', 'v1_adopted')
    ),
    CONSTRAINT platform_migration_history_path_check CHECK (
        selected_path IS NULL OR (
            pg_catalog.octet_length(selected_path) BETWEEN 1 AND 512
            AND selected_path ~ '^[a-z0-9][a-z0-9_./-]*$'
            AND pg_catalog.strpos(selected_path, '//') = 0
            AND selected_path !~ '(^|/)\.{1,2}(/|$)'
        )
    ),
    CONSTRAINT platform_migration_history_sha_check CHECK (
        selected_sha256 IS NULL OR selected_sha256 ~ '^[0-9a-f]{64}$'
    ),
    CONSTRAINT platform_migration_history_run_check CHECK (
        run_id IS NULL OR run_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT platform_migration_history_generation_check CHECK (
        (target_database_generation IS NULL OR target_database_generation > 0)
        AND (next_minimum_runtime_generation IS NULL OR next_minimum_runtime_generation > 0)
    ),
    CONSTRAINT platform_migration_history_row_shape_check CHECK (
        (
            version = 0
            AND record_kind = 'anchor'
            AND owner IS NULL
            AND scope IS NULL
            AND origin IS NULL
            AND selected_path IS NULL
            AND selected_sha256 IS NULL
            AND run_id IS NULL
            AND target_database_generation IS NULL
            AND next_minimum_runtime_generation IS NULL
        ) OR (
            version > 0
            AND record_kind = 'migration'
            AND owner IS NOT NULL
            AND scope IS NOT NULL
            AND origin IS NOT NULL
            AND selected_path IS NOT NULL
            AND selected_sha256 IS NOT NULL
            AND run_id IS NOT NULL
            AND target_database_generation IS NOT NULL
            AND next_minimum_runtime_generation IS NOT NULL
        )
    )
);
