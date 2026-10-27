-- +goose Up
CREATE TABLE platform.analytics_publications (
    public_id bytea NOT NULL,
    tenant_id uuid NOT NULL,
    release_id uuid NOT NULL,
    resource_kind text NOT NULL,
    resource_version_id uuid NOT NULL,
    audience text NOT NULL,
    state text NOT NULL,
    generation bigint NOT NULL,
    capability_sha256 bytea,
    capability_expires_at timestamptz,
    activated_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT analytics_publications_pkey PRIMARY KEY (public_id),
    CONSTRAINT analytics_publications_release_key UNIQUE (tenant_id, release_id),
    CONSTRAINT analytics_publications_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_publications_identity_check CHECK (
        pg_catalog.octet_length(public_id) = 16
        AND pg_catalog.uuid_extract_version(release_id) = 7
        AND pg_catalog.uuid_extract_version(resource_version_id) = 7
        AND resource_kind IN ('query', 'dashboard')
        AND audience IN ('public', 'capability')
        AND state IN ('active', 'revoked')
        AND generation > 0
    ),
    CONSTRAINT analytics_publications_capability_check CHECK (
        (audience = 'public' AND capability_sha256 IS NULL
         AND capability_expires_at IS NULL)
        OR (audience = 'capability'
            AND pg_catalog.octet_length(capability_sha256) = 32
            AND capability_expires_at IS NOT NULL
            AND pg_catalog.isfinite(capability_expires_at)
            AND capability_expires_at > activated_at)
    ),
    CONSTRAINT analytics_publications_time_check CHECK (
        pg_catalog.isfinite(activated_at)
        AND (revoked_at IS NULL OR pg_catalog.isfinite(revoked_at))
        AND ((state = 'active' AND revoked_at IS NULL)
          OR (state = 'revoked' AND revoked_at >= activated_at))
    )
);

CREATE TABLE platform.analytics_legacy_public_routes (
    legacy_tenant_identity text NOT NULL,
    method text NOT NULL,
    path text NOT NULL,
    slug text NOT NULL,
    tenant_id uuid NOT NULL,
    legacy_endpoint_id uuid NOT NULL,
    public_id bytea NOT NULL,
    release_id uuid NOT NULL,
    state text NOT NULL,
    hit_count bigint NOT NULL DEFAULT 0,
    last_hit_at timestamptz,
    consumers_acked_at timestamptz,
    remove_after timestamptz NOT NULL,
    CONSTRAINT analytics_legacy_public_routes_pkey PRIMARY KEY (
        legacy_tenant_identity, method, path, slug
    ),
    CONSTRAINT analytics_legacy_public_routes_endpoint_key UNIQUE (
        tenant_id, legacy_endpoint_id
    ),
    CONSTRAINT analytics_legacy_public_routes_public_fk FOREIGN KEY (public_id)
        REFERENCES platform.analytics_publications (public_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_legacy_public_routes_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT analytics_legacy_public_routes_shape_check CHECK (
        pg_catalog.octet_length(legacy_tenant_identity) BETWEEN 1 AND 200
        AND method IN ('GET', 'POST')
        AND pg_catalog.octet_length(path) BETWEEN 1 AND 512
        AND pg_catalog.octet_length(slug) BETWEEN 1 AND 100
        AND pg_catalog.uuid_extract_version(legacy_endpoint_id) IN (4, 7)
        AND pg_catalog.octet_length(public_id) = 16
        AND pg_catalog.uuid_extract_version(release_id) = 7
        AND state IN ('active', 'retired')
        AND hit_count >= 0 AND pg_catalog.isfinite(remove_after)
        AND (last_hit_at IS NULL OR pg_catalog.isfinite(last_hit_at))
        AND (consumers_acked_at IS NULL OR pg_catalog.isfinite(consumers_acked_at))
    )
);

CREATE TABLE platform.analytics_public_rate_windows (
    tenant_id uuid NOT NULL,
    release_id uuid NOT NULL,
    generation bigint NOT NULL,
    client_ip inet NOT NULL,
    window_started_at timestamptz NOT NULL,
    request_count integer NOT NULL,
    CONSTRAINT analytics_public_rate_windows_pkey PRIMARY KEY (
        tenant_id, release_id, generation, client_ip, window_started_at
    ),
    CONSTRAINT analytics_public_rate_windows_shape_check CHECK (
        generation > 0 AND request_count BETWEEN 1 AND 60
        AND pg_catalog.isfinite(window_started_at)
    )
);

CREATE TABLE platform.analytics_public_concurrency_leases (
    lease_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    release_id uuid NOT NULL,
    generation bigint NOT NULL,
    expires_at timestamptz NOT NULL,
    CONSTRAINT analytics_public_concurrency_leases_pkey PRIMARY KEY (lease_id),
    CONSTRAINT analytics_public_concurrency_release_index UNIQUE (
        release_id, lease_id
    ),
    CONSTRAINT analytics_public_concurrency_leases_shape_check CHECK (
        pg_catalog.uuid_extract_version(lease_id) = 7
        AND generation > 0 AND pg_catalog.isfinite(expires_at)
    )
);

