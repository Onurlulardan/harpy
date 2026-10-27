-- +goose Up
CREATE TABLE iam_profile_fields (
    id uuid NOT NULL,
    resource_kind text NOT NULL,
    field_key text COLLATE "C" NOT NULL,
    value_kind text NOT NULL,
    cardinality text NOT NULL,
    required boolean NOT NULL,
    classification text NOT NULL,
    normalization_profile text NOT NULL,
    state text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    retired_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT iam_profile_fields_pkey PRIMARY KEY (id),
    CONSTRAINT iam_profile_fields_resource_field_key
        UNIQUE (resource_kind, field_key),
    CONSTRAINT iam_profile_fields_resource_id_key UNIQUE (resource_kind, id),
    CONSTRAINT iam_profile_fields_resource_kind_check
        CHECK (resource_kind IN ('user', 'organization', 'role')),
    CONSTRAINT iam_profile_fields_field_key_check CHECK (
        pg_catalog.char_length(field_key) <= 63
        AND field_key ~ '^[a-z][a-z0-9]*(_[a-z0-9]+)*$'
    ),
    CONSTRAINT iam_profile_fields_value_kind_check CHECK (
        value_kind IN (
            'string', 'integer', 'decimal', 'boolean', 'date',
            'timestamp', 'uuid', 'enum', 'json'
        )
    ),
    CONSTRAINT iam_profile_fields_cardinality_check
        CHECK (cardinality IN ('single', 'list')),
    CONSTRAINT iam_profile_fields_classification_check
        CHECK (classification IN ('public', 'internal', 'confidential', 'restricted')),
    CONSTRAINT iam_profile_fields_normalization_profile_check CHECK (
        normalization_profile IN (
            'none_v1', 'email_v1', 'portal_v1'
        )
    ),
    CONSTRAINT iam_profile_fields_normalization_type_check CHECK (
        normalization_profile = 'none_v1' OR value_kind = 'string'
    ),
    CONSTRAINT iam_profile_fields_state_check
        CHECK (state IN ('active', 'retired')),
    CONSTRAINT iam_profile_fields_revision_check CHECK (revision > 0),
    CONSTRAINT iam_profile_fields_timestamp_order_check CHECK (
        updated_at >= created_at AND
        (retired_at IS NULL OR retired_at >= updated_at)
    ),
    CONSTRAINT iam_profile_fields_state_timestamps_check CHECK (
        (state = 'active' AND retired_at IS NULL)
        OR (state = 'retired' AND retired_at IS NOT NULL)
    )
);
