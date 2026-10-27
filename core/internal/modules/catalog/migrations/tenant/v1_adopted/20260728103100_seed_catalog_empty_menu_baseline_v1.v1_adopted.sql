-- +goose Up
-- Migration stream version: 20260728103100.

INSERT INTO catalog_menu_definitions (
    definition_id, revision, active_version_id, updated_at
)
VALUES (
    '018f6000-0000-7000-8000-000000000001'::uuid,
    1, NULL, pg_catalog.clock_timestamp()
)
ON CONFLICT (definition_id) DO NOTHING;

INSERT INTO catalog_menu_versions (
    version_id, definition_id, version_number,
    content_sha256, created_at
)
SELECT
    '018f6000-0000-7000-8000-000000000002'::uuid,
    definition.definition_id,
    1,
    'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
    pg_catalog.clock_timestamp()
FROM catalog_menu_definitions AS definition
WHERE definition.definition_id =
      '018f6000-0000-7000-8000-000000000001'::uuid
  AND definition.active_version_id IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM catalog_menu_versions AS existing
      WHERE existing.definition_id = definition.definition_id
  )
ON CONFLICT (version_id) DO NOTHING;

UPDATE catalog_menu_definitions
SET active_version_id = '018f6000-0000-7000-8000-000000000002'::uuid,
    updated_at = pg_catalog.clock_timestamp()
WHERE definition_id = '018f6000-0000-7000-8000-000000000001'::uuid
  AND active_version_id IS NULL;

SELECT 1 / CASE WHEN NOT EXISTS (
    SELECT 1
    FROM catalog_menu_definitions AS definition
    JOIN catalog_menu_versions AS version
      ON version.version_id = definition.active_version_id
     AND version.definition_id = definition.definition_id
    WHERE definition.definition_id =
          '018f6000-0000-7000-8000-000000000001'::uuid
      AND definition.revision >= 1
      AND version.version_number >= 1
) OR EXISTS (
    SELECT 1
    FROM catalog_menu_definitions AS definition
    JOIN catalog_menu_versions AS version
      ON version.version_id = definition.active_version_id
    WHERE definition.definition_id =
          '018f6000-0000-7000-8000-000000000001'::uuid
      AND definition.active_version_id =
          '018f6000-0000-7000-8000-000000000002'::uuid
      AND (
          version.version_number <> 1
          OR version.content_sha256 <>
             'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'
          OR EXISTS (
              SELECT 1
              FROM catalog_menu_items AS item
              WHERE item.version_id = version.version_id
          )
      )
) THEN 0 ELSE 1 END;
