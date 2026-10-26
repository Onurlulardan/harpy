SELECT id, name, slug, status, settings_revision, created_at, updated_at
FROM platform.tenancy_settings_read_v1($1::uuid);
