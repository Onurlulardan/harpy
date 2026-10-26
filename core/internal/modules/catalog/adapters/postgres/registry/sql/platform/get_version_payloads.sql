SELECT version.package_id::text, package.slug, version.version,
       version.manifest, version.manifest_sha256, version.signature,
       version.key_id, version.release_id::text, version.registered_at,
       payloads.documents
FROM platform.catalog_package_versions AS version
JOIN platform.catalog_packages AS package ON package.package_id = version.package_id
CROSS JOIN LATERAL (
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'component_id', payload.component_id::text,
                'owner', payload.owner,
                'contract', payload.contract,
                'content_sha256', payload.content_sha256,
                'document', payload.document,
                'release_id', payload.release_id::text,
                'registered_at', payload.registered_at
            )
            ORDER BY payload.component_id::text
        ),
        '[]'::jsonb
    ) AS documents
    FROM platform.catalog_package_owner_payloads AS payload
    WHERE payload.package_id = version.package_id
      AND payload.version = version.version
) AS payloads
WHERE version.package_id = $1::uuid AND version.version = $2::text
