-- +goose Up
ALTER TABLE iam_profile_field_revisions
    ALTER COLUMN field_key SET NOT NULL,
    ALTER COLUMN value_kind SET NOT NULL,
    ALTER COLUMN cardinality SET NOT NULL,
    ALTER COLUMN required SET NOT NULL,
    ALTER COLUMN classification SET NOT NULL,
    ALTER COLUMN normalization_profile SET NOT NULL,
    ALTER COLUMN field_state SET NOT NULL,
    ALTER COLUMN semantic_sha256 SET NOT NULL,
    ADD CONSTRAINT iam_profile_field_revisions_field_key_check CHECK (
        pg_catalog.char_length(field_key) <= 63
        AND field_key ~ '^[a-z][a-z0-9]*(_[a-z0-9]+)*$'
    ),
    ADD CONSTRAINT iam_profile_field_revisions_value_kind_check CHECK (
        value_kind IN (
            'string', 'integer', 'decimal', 'boolean', 'date',
            'timestamp', 'uuid', 'enum', 'json'
        )
    ),
    ADD CONSTRAINT iam_profile_field_revisions_cardinality_check
        CHECK (cardinality IN ('single', 'list')),
    ADD CONSTRAINT iam_profile_field_revisions_classification_check CHECK (
        classification IN ('public', 'internal', 'confidential', 'restricted')
    ),
    ADD CONSTRAINT iam_profile_field_revisions_normalization_check CHECK (
        normalization_profile IN ('none_v1', 'email_v1', 'portal_v1')
        AND (normalization_profile = 'none_v1' OR value_kind = 'string')
    ),
    ADD CONSTRAINT iam_profile_field_revisions_field_state_check
        CHECK (field_state IN ('active', 'retired')),
    ADD CONSTRAINT iam_profile_field_revisions_semantic_sha256_check CHECK (
        pg_catalog.octet_length(semantic_sha256) = 32
        AND semantic_sha256 = iam_profile_semantic_sha256_v1(
            field_id, resource_kind, field_key, value_kind, cardinality,
            required, classification, normalization_profile
        )
    ),
    ADD CONSTRAINT iam_profile_field_revisions_validation_digest_check CHECK (
        validation_sha256 = pg_catalog.sha256(pg_catalog.convert_to(
            iam_profile_validation_canonical_v1(validation_schema), 'UTF8'
        ))
    );
