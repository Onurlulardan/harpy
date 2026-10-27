WITH payload_input AS MATERIALIZED (
    SELECT component_id::uuid, owner, contract, content_sha256, document
    FROM jsonb_to_recordset($10::jsonb) AS input(
        component_id text,
        owner text,
        contract text,
        content_sha256 text,
        document jsonb
    )
),
inserted_package AS (
    INSERT INTO platform.catalog_packages (package_id, slug, created_at)
    VALUES ($1::uuid, $2::text, $9::timestamptz)
    ON CONFLICT DO NOTHING
    RETURNING package_id, slug
),
resolved_package AS MATERIALIZED (
    SELECT package_id, slug FROM inserted_package
    UNION ALL
    SELECT package_id, slug
    FROM platform.catalog_packages
    WHERE package_id = $1::uuid AND slug = $2::text
    LIMIT 1
),
inserted_version AS (
    INSERT INTO platform.catalog_package_versions (
        package_id, version, manifest, manifest_sha256, signature,
        key_id, release_id, registered_at
    )
    SELECT package_id, $3::text, $4::jsonb, $5::text, $6::bytea,
           $7::text, $8::uuid, $9::timestamptz
    FROM resolved_package
    ON CONFLICT DO NOTHING
    RETURNING package_id, version, manifest, manifest_sha256, signature,
              key_id, release_id, registered_at
),
resolved_version AS MATERIALIZED (
    SELECT package_id, version, manifest, manifest_sha256, signature,
           key_id, release_id, registered_at
    FROM inserted_version
    UNION ALL
    SELECT package_id, version, manifest, manifest_sha256, signature,
           key_id, release_id, registered_at
    FROM platform.catalog_package_versions
    WHERE package_id = $1::uuid AND version = $3::text
    LIMIT 1
),
inserted_payloads AS (
    INSERT INTO platform.catalog_package_owner_payloads (
        package_id, version, component_id, owner, contract,
        content_sha256, document, release_id, registered_at
    )
    SELECT version.package_id, version.version, payload.component_id,
           payload.owner, payload.contract, payload.content_sha256,
           payload.document, version.release_id, version.registered_at
    FROM resolved_version AS version
    CROSS JOIN payload_input AS payload
    ON CONFLICT DO NOTHING
    RETURNING package_id, version, component_id, owner, contract,
              content_sha256, document, release_id, registered_at
),
resolved_payloads AS MATERIALIZED (
    SELECT package_id, version, component_id, owner, contract,
           content_sha256, document, release_id, registered_at
    FROM inserted_payloads
    UNION ALL
    SELECT package_id, version, component_id, owner, contract,
           content_sha256, document, release_id, registered_at
    FROM platform.catalog_package_owner_payloads
    WHERE package_id = $1::uuid AND version = $3::text
),
encoded_payloads AS (
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'component_id', component_id::text,
                'owner', owner,
                'contract', contract,
                'content_sha256', content_sha256,
                'document', document,
                'release_id', release_id::text,
                'registered_at', registered_at
            )
            ORDER BY component_id::text
        ),
        '[]'::jsonb
    ) AS payloads
    FROM resolved_payloads
)
SELECT version.package_id::text, package.slug, version.version,
       version.manifest, version.manifest_sha256, version.signature,
       version.key_id, version.release_id::text, version.registered_at,
       payloads.payloads
FROM resolved_version AS version
JOIN resolved_package AS package ON package.package_id = version.package_id
CROSS JOIN encoded_payloads AS payloads
