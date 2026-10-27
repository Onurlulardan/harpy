-- +goose Up
CREATE TABLE platform.platformops_trace_spans (
    id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    correlation_id text NOT NULL CHECK (
        length(correlation_id) BETWEEN 1 AND 64 AND
        correlation_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$'
    ),
    span_id text NOT NULL CHECK (span_id ~ '^[0-9a-f]{16,32}$'),
    parent_span_id text CHECK (parent_span_id IS NULL OR parent_span_id ~ '^[0-9a-f]{16,32}$'),
    service_name text NOT NULL CHECK (length(service_name) BETWEEN 1 AND 100),
    operation text NOT NULL CHECK (length(operation) BETWEEN 1 AND 256),
    span_kind text NOT NULL CHECK (span_kind IN ('server','client','internal','consumer','producer')),
    method text NOT NULL CHECK (method IN ('GET','HEAD','POST','PUT','PATCH','DELETE','OPTIONS')),
    path text NOT NULL CHECK (length(path) BETWEEN 1 AND 256 AND path LIKE '/%'),
    status_code integer NOT NULL CHECK (status_code BETWEEN 100 AND 599),
    started_at timestamptz NOT NULL,
    ended_at timestamptz NOT NULL CHECK (ended_at >= started_at),
    duration_ms numeric(18,3) NOT NULL CHECK (duration_ms >= 0),
    level text NOT NULL CHECK (level IN ('debug','info','warn','error','fatal')),
    message text CHECK (message IS NULL OR length(message) <= 500),
    created_at timestamptz NOT NULL,
    UNIQUE (environment_id, span_id)
);

CREATE INDEX platformops_trace_correlation_idx
ON platform.platformops_trace_spans(environment_id, correlation_id, started_at);
CREATE INDEX platformops_trace_time_idx
ON platform.platformops_trace_spans(environment_id, started_at DESC);
CREATE INDEX platformops_trace_filter_idx
ON platform.platformops_trace_spans(environment_id, service_name, level, started_at DESC);

CREATE TABLE platform.platformops_alert_rules (
    id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 120),
    description text NOT NULL CHECK (length(description) <= 1000),
    enabled boolean NOT NULL,
    service_name text CHECK (service_name IS NULL OR length(service_name) BETWEEN 1 AND 100),
    condition text NOT NULL CHECK (condition IN ('unhealthy','degraded','any_change')),
    cooldown_minutes integer NOT NULL CHECK (cooldown_minutes BETWEEN 1 AND 10080),
    channels jsonb NOT NULL CHECK (jsonb_typeof(channels) = 'array' AND jsonb_array_length(channels) <= 8),
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (environment_id, name)
);

CREATE TABLE platform.platformops_alert_events (
    id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    alert_rule_id uuid,
    service_name text NOT NULL CHECK (length(service_name) BETWEEN 1 AND 100),
    previous_status text NOT NULL CHECK (previous_status IN ('healthy','degraded','unhealthy','unknown')),
    current_status text NOT NULL CHECK (current_status IN ('healthy','degraded','unhealthy','unknown')),
    channels_used text[] NOT NULL DEFAULT '{}',
    message text NOT NULL CHECK (length(message) BETWEEN 1 AND 1000),
    actor_id uuid,
    resolved_at timestamptz,
    created_at timestamptz NOT NULL
);

CREATE INDEX platformops_alert_events_time_idx
ON platform.platformops_alert_events(environment_id, created_at DESC);
CREATE INDEX platformops_alert_events_filter_idx
ON platform.platformops_alert_events(environment_id, service_name, current_status, created_at DESC);
