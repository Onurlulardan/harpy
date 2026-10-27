SELECT version.manifest_sha256,version.row_revision
FROM analytics_saved_query_versions AS version
WHERE version.query_id=$1::uuid AND version.id=$2::uuid
  AND version.manifest_sha256=$3::bytea
  AND version.version_state='published';
