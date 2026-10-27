-- +goose Up
CREATE TABLE notifications_preferences (
    recipient_id uuid NOT NULL,
    category text COLLATE "C" NOT NULL CHECK (category IN (
        'security','workflow_action','billing','document','system','product'
    )),
    channel text COLLATE "C" NOT NULL CHECK (channel IN ('in_app','email','telegram')),
    enabled boolean NOT NULL,
    explicit boolean NOT NULL,
    verified_target boolean NOT NULL,
    quiet_timezone text COLLATE "C",
    quiet_start_minute integer,
    quiet_end_minute integer,
    row_version bigint NOT NULL CHECK (row_version > 0),
    updated_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(updated_at)),
    PRIMARY KEY (recipient_id, category, channel),
    CHECK (
        (quiet_timezone IS NULL AND quiet_start_minute IS NULL AND quiet_end_minute IS NULL)
        OR (
            channel <> 'in_app'
            AND pg_catalog.octet_length(quiet_timezone) BETWEEN 1 AND 128
            AND quiet_start_minute BETWEEN 0 AND 1439
            AND quiet_end_minute BETWEEN 0 AND 1439
            AND quiet_start_minute <> quiet_end_minute
        )
    ),
    CHECK (NOT (
        enabled = false AND (
            (category <> 'product' AND channel = 'in_app')
            OR (category = 'security' AND channel = 'email')
        )
    )),
    CHECK (channel <> 'telegram' OR enabled = false OR (explicit AND verified_target))
);

CREATE INDEX notifications_preferences_recipient_idx
    ON notifications_preferences (recipient_id, category, channel);
