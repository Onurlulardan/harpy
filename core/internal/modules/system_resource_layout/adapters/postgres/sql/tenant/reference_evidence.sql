SELECT live_reference_count, manifest_sha256, checked_owner_count
FROM system_resource_layout_field_reference_evidence_v2(
    (pg_catalog.current_schemas(false))[2]::name, $1::text, $2::uuid
);
