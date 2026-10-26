SELECT owner_domain, resource_kind, resource_id, resource_version_id,
       owner_generation, catalog_version, query_contract_version,
       policy_version, data_revision_capability, dependency_sha256
FROM analytics_query_dependencies_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1::uuid, $2::uuid
);
