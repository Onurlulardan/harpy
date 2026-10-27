-- +goose Up
CREATE TABLE iam_portal_subject_bindings (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    access_mode_id uuid NOT NULL,
    access_mode text NOT NULL,
    subject_shape text NOT NULL,
    subject_type text COLLATE "C" NOT NULL,
    profile_owner text NOT NULL,
    profile_form_id uuid,
    profile_form_version_id uuid,
    subject_record_id uuid NOT NULL,
    identifier_field_id uuid,
    contract_sha256 bytea NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    activated_at timestamptz,
    suspended_at timestamptz,
    retired_at timestamptz,
    CONSTRAINT iam_portal_subject_bindings_pkey PRIMARY KEY (id),
    CONSTRAINT iam_portal_subject_bindings_user_key UNIQUE (id, user_id),
    CONSTRAINT iam_portal_subject_bindings_setup_key
        UNIQUE (id, user_id, access_mode_id),
    CONSTRAINT iam_portal_subject_bindings_access_mode_fk
        FOREIGN KEY (access_mode_id, user_id, access_mode)
        REFERENCES iam_user_access_modes (id, user_id, mode)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_portal_subject_bindings_access_mode_check
        CHECK (access_mode = 'portal'),
    CONSTRAINT iam_portal_subject_bindings_subject_shape_check
        CHECK (subject_shape IN ('self', 'form_record')),
    CONSTRAINT iam_portal_subject_bindings_subject_type_check CHECK (
        pg_catalog.char_length(subject_type) BETWEEN 1 AND 100 AND
        subject_type ~ '^[a-z0-9_-]+$'
    ),
    CONSTRAINT iam_portal_subject_bindings_profile_owner_check
        CHECK (profile_owner IN ('iam', 'forms')),
    CONSTRAINT iam_portal_subject_bindings_subject_shape_fields_check CHECK (
        (
            subject_shape = 'self' AND
            subject_type = 'operator' AND
            profile_owner = 'iam' AND
            profile_form_id IS NULL AND
            profile_form_version_id IS NULL AND
            subject_record_id = user_id AND
            identifier_field_id IS NULL
        ) OR
        (
            subject_shape = 'form_record' AND
            profile_owner = 'forms' AND
            profile_form_id IS NOT NULL AND
            profile_form_version_id IS NOT NULL AND
            identifier_field_id IS NOT NULL
        )
    ),
    CONSTRAINT iam_portal_subject_bindings_contract_hash_check
        CHECK (pg_catalog.octet_length(contract_sha256) = 32),
    CONSTRAINT iam_portal_subject_bindings_state_check
        CHECK (state IN ('pending', 'active', 'suspended', 'retired')),
    CONSTRAINT iam_portal_subject_bindings_revision_check CHECK (revision > 0),
    CONSTRAINT iam_portal_subject_bindings_timestamp_check CHECK (
        updated_at >= created_at AND
        (activated_at IS NULL OR
            (activated_at >= created_at AND activated_at <= updated_at)) AND
        (suspended_at IS NULL OR
            (suspended_at >= COALESCE(activated_at, created_at) AND
                suspended_at <= updated_at)) AND
        (retired_at IS NULL OR
            (retired_at >= created_at AND retired_at <= updated_at))
    ),
    CONSTRAINT iam_portal_subject_bindings_lifecycle_check CHECK (
        (state = 'pending' AND activated_at IS NULL AND
            suspended_at IS NULL AND retired_at IS NULL) OR
        (state = 'active' AND activated_at IS NOT NULL AND
            suspended_at IS NULL AND retired_at IS NULL) OR
        (state = 'suspended' AND suspended_at IS NOT NULL AND
            retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND
            (suspended_at IS NULL OR retired_at >= suspended_at) AND
            (activated_at IS NULL OR retired_at >= activated_at))
    )
);

CREATE UNIQUE INDEX iam_portal_subject_bindings_live_user_uidx
    ON iam_portal_subject_bindings (user_id)
    WHERE state <> 'retired';

CREATE UNIQUE INDEX iam_portal_subject_bindings_live_form_record_uidx
    ON iam_portal_subject_bindings (profile_form_id, subject_record_id)
    WHERE subject_shape = 'form_record' AND state <> 'retired';

CREATE INDEX iam_portal_subject_bindings_profile_idx
    ON iam_portal_subject_bindings (
        profile_owner, profile_form_id, profile_form_version_id
    )
    WHERE state <> 'retired';
