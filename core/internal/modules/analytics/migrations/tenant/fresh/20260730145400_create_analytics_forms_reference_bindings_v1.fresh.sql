-- +goose Up
CREATE TABLE analytics_forms_reference_bindings (
    reference_id uuid PRIMARY KEY,
    register_operation_id uuid NOT NULL UNIQUE,
    source_kind text NOT NULL,
    source_id uuid NOT NULL,
    source_version_id uuid NOT NULL,
    component_id uuid,
    target_definition_id uuid NOT NULL,
    target_version_id uuid NOT NULL,
    usage text NOT NULL DEFAULT 'READ',
    claimed_by uuid NOT NULL,
    claimed_at timestamptz NOT NULL,
    target_reference_epoch bigint,
    register_result_sha256 bytea,
    registered_at timestamptz,
    release_operation_id uuid UNIQUE,
    release_claimed_at timestamptz,
    release_result_sha256 bytea,
    released_at timestamptz,
    CONSTRAINT analytics_forms_reference_source_key UNIQUE (
        source_id,source_version_id,target_definition_id,target_version_id,usage
    ),
    CONSTRAINT analytics_forms_reference_identity_check CHECK (
        pg_catalog.uuid_extract_version(reference_id)=7 AND
        pg_catalog.uuid_extract_version(register_operation_id)=7 AND
        pg_catalog.uuid_extract_version(source_id)=7 AND
        pg_catalog.uuid_extract_version(source_version_id)=7 AND
        pg_catalog.uuid_extract_version(target_definition_id)=7 AND
        pg_catalog.uuid_extract_version(target_version_id)=7 AND
        claimed_by<>'00000000-0000-0000-0000-000000000000'::uuid AND
        source_kind IN ('query','catalog_query') AND usage='READ' AND
        ((source_kind='query' AND component_id IS NULL) OR
         (source_kind='catalog_query' AND
          pg_catalog.uuid_extract_version(component_id)=7)) AND
        pg_catalog.isfinite(claimed_at) AND
        date_trunc('microseconds',claimed_at)=claimed_at
    ),
    CONSTRAINT analytics_forms_reference_register_check CHECK (
        (target_reference_epoch IS NULL AND register_result_sha256 IS NULL
         AND registered_at IS NULL) OR
        (target_reference_epoch>0 AND
         pg_catalog.octet_length(register_result_sha256)=32 AND
         registered_at=claimed_at)
    ),
    CONSTRAINT analytics_forms_reference_release_check CHECK (
        (release_operation_id IS NULL AND release_claimed_at IS NULL
         AND release_result_sha256 IS NULL AND released_at IS NULL) OR
        (pg_catalog.uuid_extract_version(release_operation_id)=7 AND
         release_claimed_at>=claimed_at AND
         pg_catalog.isfinite(release_claimed_at) AND
         date_trunc('microseconds',release_claimed_at)=release_claimed_at AND
         release_result_sha256 IS NULL AND released_at IS NULL) OR
        (pg_catalog.uuid_extract_version(release_operation_id)=7 AND
         release_claimed_at>=claimed_at AND
         pg_catalog.octet_length(release_result_sha256)=32 AND
         released_at=release_claimed_at)
    )
);
