-- +goose Up
-- +goose StatementBegin
CREATE TABLE platform.usage_request_facts (
    fact_id uuid PRIMARY KEY CHECK (pg_catalog.uuid_extract_version(fact_id) = 7),
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    source_identity bytea NOT NULL CHECK (pg_catalog.octet_length(source_identity) = 32),
    fact_digest bytea NOT NULL CHECK (pg_catalog.octet_length(fact_digest) = 32),
    method text NOT NULL CHECK (method IN ('GET','HEAD','POST','PUT','PATCH','DELETE','OPTIONS')),
    route_pattern text NOT NULL CHECK (
        pg_catalog.octet_length(route_pattern) BETWEEN 1 AND 256
        AND pg_catalog.left(route_pattern, 1) = '/'
        AND pg_catalog.strpos(route_pattern, '?') = 0
        AND pg_catalog.strpos(route_pattern, '#') = 0
        AND route_pattern !~ E'[\r\n]'
    ),
    status_code integer NOT NULL CHECK (status_code BETWEEN 100 AND 599),
    duration_micros bigint NOT NULL CHECK (duration_micros BETWEEN 0 AND 86400000000),
    occurred_at timestamp(6) with time zone NOT NULL,
    created_at timestamp(6) with time zone NOT NULL,
    UNIQUE (environment_id, source_identity)
);

CREATE INDEX usage_request_facts_tenant_time_idx
ON platform.usage_request_facts (environment_id, tenant_id, occurred_at DESC, fact_id DESC);
CREATE INDEX usage_request_facts_actor_time_idx
ON platform.usage_request_facts (environment_id, tenant_id, actor_id, occurred_at DESC, fact_id DESC);
CREATE INDEX usage_request_facts_route_time_idx
ON platform.usage_request_facts (environment_id, tenant_id, route_pattern, method, occurred_at DESC);

CREATE FUNCTION platform.usage_request_fact_immutable_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
BEGIN
    IF TG_OP = 'DELETE'
       AND pg_catalog.current_setting('forgecrud.usage_retention', true) = 'usage_retention_v1' THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION USING ERRCODE = 'P6903', MESSAGE = 'usage facts are immutable';
END;
$$;

CREATE TRIGGER usage_request_facts_immutable
BEFORE UPDATE OR DELETE ON platform.usage_request_facts
FOR EACH ROW EXECUTE FUNCTION platform.usage_request_fact_immutable_v1();

-- The table is owner-only on creation. Runtime and adopter roles receive
-- routine EXECUTE privileges only; no direct relation privilege is granted.
REVOKE ALL ON FUNCTION platform.usage_request_fact_immutable_v1() FROM PUBLIC;
-- +goose StatementEnd
