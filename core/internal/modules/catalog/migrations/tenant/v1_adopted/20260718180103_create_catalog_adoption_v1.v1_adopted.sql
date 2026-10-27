-- +goose Up
CREATE TABLE catalog_adoption_runs (
    run_id uuid PRIMARY KEY,
    source_manifest_sha256 text NOT NULL CHECK (source_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    transform_sha256 text NOT NULL CHECK (transform_sha256 ~ '^[0-9a-f]{64}$'),
    source_count bigint NOT NULL CHECK (source_count >= 0),
    target_count bigint NOT NULL CHECK (target_count >= 0),
    source_root_sha256 text NOT NULL CHECK (source_root_sha256 ~ '^[0-9a-f]{64}$'),
    target_root_sha256 text NOT NULL CHECK (target_root_sha256 ~ '^[0-9a-f]{64}$'),
    state text NOT NULL CHECK (state IN ('INACTIVE','VERIFIED','BLOCKED','RETIRED')),
    blocker_code text,
    cutover_generation bigint NOT NULL CHECK (cutover_generation > 0),
    created_at timestamptz NOT NULL,
    verified_at timestamptz,
	CHECK (blocker_code IS NULL OR length(blocker_code) BETWEEN 1 AND 2048),
    CHECK (date_trunc('microseconds', created_at) = created_at),
    CHECK (verified_at IS NULL OR date_trunc('microseconds', verified_at) = verified_at),
    CHECK ((state = 'BLOCKED') = (blocker_code IS NOT NULL))
);

CREATE TABLE catalog_legacy_route_hits (
    route_key text PRIMARY KEY CHECK (length(route_key) BETWEEN 1 AND 256),
    hit_count bigint NOT NULL CHECK (hit_count >= 0),
    last_hit_at timestamptz,
    consumer_ack_count bigint NOT NULL CHECK (consumer_ack_count >= 0),
    required_consumer_count bigint NOT NULL CHECK (required_consumer_count >= 0),
    remove_after timestamptz NOT NULL,
    retired_at timestamptz,
    CHECK (last_hit_at IS NULL OR date_trunc('microseconds', last_hit_at) = last_hit_at),
    CHECK (date_trunc('microseconds', remove_after) = remove_after),
    CHECK (retired_at IS NULL OR date_trunc('microseconds', retired_at) = retired_at)
);

CREATE TABLE catalog_legacy_consumer_acks (
    route_key text NOT NULL,
    consumer_id text NOT NULL,
    evidence_sha256 text NOT NULL CHECK (evidence_sha256 ~ '^[0-9a-f]{64}$'),
    acknowledged_at timestamptz NOT NULL,
    PRIMARY KEY (route_key, consumer_id),
    CONSTRAINT catalog_legacy_consumer_acks_route_fk
        FOREIGN KEY (route_key) REFERENCES catalog_legacy_route_hits(route_key)
        ON DELETE RESTRICT,
    CHECK (length(consumer_id) BETWEEN 3 AND 128),
    CHECK (date_trunc('microseconds', acknowledged_at) = acknowledged_at)
);

CREATE TABLE catalog_tombstones (
    package_id uuid PRIMARY KEY,
    slug text NOT NULL,
    package_sha256 text NOT NULL CHECK (package_sha256 ~ '^[0-9a-f]{64}$'),
    reference_epoch bigint NOT NULL CHECK (reference_epoch > 0),
    owner_receipt_root text NOT NULL CHECK (owner_receipt_root ~ '^[0-9a-f]{64}$'),
    purged_at timestamptz NOT NULL,
    CHECK (date_trunc('microseconds', purged_at) = purged_at)
);
