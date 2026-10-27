-- +goose Up
CREATE TABLE platform.notification_realtime_ticket_audits (
    ticket_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    authorization_version bigint NOT NULL CHECK (authorization_version > 0),
    origin_digest bytea NOT NULL CHECK (pg_catalog.octet_length(origin_digest) = 32),
    terminal_kind text COLLATE "C" NOT NULL CHECK (terminal_kind IN ('consumed','expired')),
    terminal_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(terminal_at)),
    retain_until timestamptz NOT NULL CHECK (pg_catalog.isfinite(retain_until)),
    CHECK (retain_until = terminal_at + interval '90 days')
);

CREATE INDEX notification_realtime_ticket_audits_retention_idx
    ON platform.notification_realtime_ticket_audits (retain_until, ticket_id);
CREATE INDEX notification_legacy_connection_hits_purge_idx
    ON platform.notification_legacy_connection_hits (occurred_at, id);

GRANT SELECT ON TABLE platform.notification_realtime_ticket_audits TO forgecrud_runtime;
GRANT INSERT (ticket_id, tenant_id, user_id, authorization_version, origin_digest, terminal_kind, terminal_at, retain_until) ON platform.notification_realtime_ticket_audits TO forgecrud_runtime;
