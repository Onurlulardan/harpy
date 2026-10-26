SELECT version.manifest_sha256,
       pg_catalog.octet_length(version.query_document)
         + pg_catalog.octet_length(version.parameter_document)
         + pg_catalog.octet_length(version.output_document) AS document_bytes,
       pg_catalog.count(dependency.query_version_id) AS dependency_count
FROM analytics_saved_query_versions AS version
LEFT JOIN analytics_query_dependencies AS dependency
  ON dependency.query_version_id=version.id
WHERE version.query_id=$1::uuid AND version.id=$2::uuid
  AND version.manifest_sha256=$3::bytea
  AND version.version_state IN ('published','needs_revalidation')
GROUP BY version.id,version.manifest_sha256,version.query_document,
         version.parameter_document,version.output_document;
