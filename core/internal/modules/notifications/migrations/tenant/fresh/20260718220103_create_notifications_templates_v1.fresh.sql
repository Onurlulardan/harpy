-- +goose Up
CREATE TABLE notifications_templates (
    id uuid PRIMARY KEY,
    notification_type text COLLATE "C" NOT NULL
        CHECK (pg_catalog.octet_length(notification_type) BETWEEN 1 AND 96),
    published_version_id uuid,
    row_version bigint NOT NULL CHECK (row_version > 0),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    updated_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(updated_at)),
    CONSTRAINT notifications_templates_type_key UNIQUE (notification_type)
);

CREATE TABLE notifications_template_versions (
    id uuid PRIMARY KEY,
    template_id uuid NOT NULL,
    version_number bigint NOT NULL CHECK (version_number > 0),
    locale text COLLATE "C" NOT NULL CHECK (pg_catalog.octet_length(locale) BETWEEN 2 AND 16),
    template_kind text COLLATE "C" NOT NULL CHECK (template_kind IN ('system_html','plain_text')),
    subject_pattern text NOT NULL CHECK (pg_catalog.octet_length(subject_pattern) BETWEEN 1 AND 4096),
    body_pattern text NOT NULL CHECK (pg_catalog.octet_length(body_pattern) BETWEEN 1 AND 65536),
    variables jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(variables) = 'array'),
    mandatory_email boolean NOT NULL,
    content_digest bytea NOT NULL CHECK (pg_catalog.octet_length(content_digest) = 32),
    published_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(published_at)),
    CONSTRAINT notifications_template_versions_release_key
        UNIQUE (template_id, version_number, locale),
    CONSTRAINT notifications_template_versions_template_fk FOREIGN KEY (template_id)
        REFERENCES notifications_templates(id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

ALTER TABLE notifications_templates
    ADD CONSTRAINT notifications_templates_published_version_fk
    FOREIGN KEY (published_version_id) REFERENCES notifications_template_versions(id)
    ON UPDATE NO ACTION ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE notifications_messages
    ADD CONSTRAINT notifications_messages_template_version_fk
    FOREIGN KEY (template_version_id) REFERENCES notifications_template_versions(id)
    ON UPDATE NO ACTION ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;

CREATE INDEX notifications_template_versions_lookup_idx
    ON notifications_template_versions (template_id, locale, version_number DESC);

INSERT INTO notifications_templates (
    id, notification_type, published_version_id, row_version, created_at, updated_at
) VALUES (
    '12000000-0000-4000-8000-000000000001', 'legacy.v1.notification', NULL, 1,
    '2026-07-18T00:00:00Z', '2026-07-18T00:00:00Z'
);

INSERT INTO notifications_template_versions (
    id, template_id, version_number, locale, template_kind, subject_pattern,
    body_pattern, variables, mandatory_email, content_digest, published_at
) VALUES (
    '12000000-0000-4000-8000-000000000002',
    '12000000-0000-4000-8000-000000000001', 1, 'tr', 'plain_text',
    'Legacy V1 notification', 'Adopted legacy content', '[]'::jsonb, false,
    decode('5c1d68f24e256ebce6665f387de27e96094c74e3d5fadb7229f2eb9f4d7ca0fd', 'hex'),
    '2026-07-18T00:00:00Z'
);

UPDATE notifications_templates
SET published_version_id = '12000000-0000-4000-8000-000000000002'
WHERE id = '12000000-0000-4000-8000-000000000001';
