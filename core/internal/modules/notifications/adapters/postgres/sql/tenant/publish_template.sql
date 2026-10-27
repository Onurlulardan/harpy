UPDATE notifications_templates template
SET published_version_id = $2::uuid,
    row_version = row_version + 1,
    updated_at = pg_catalog.transaction_timestamp()
WHERE template.id = $1::uuid AND template.row_version = $3
  AND EXISTS (
      SELECT 1 FROM notifications_template_versions version
      WHERE version.id = $2::uuid AND version.template_id = template.id
  );
