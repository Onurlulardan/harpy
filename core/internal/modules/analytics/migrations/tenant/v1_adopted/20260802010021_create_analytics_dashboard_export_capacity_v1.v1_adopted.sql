-- +goose Up
CREATE TABLE analytics_dashboard_export_capacity_counters (
    tenant_id uuid PRIMARY KEY,
    admission_count bigint NOT NULL DEFAULT 0,
    running_count bigint NOT NULL DEFAULT 0,
    revision bigint NOT NULL DEFAULT 1,
    updated_at timestamptz NOT NULL,
    CHECK(tenant_id<>'00000000-0000-0000-0000-000000000000'::uuid),
    CHECK(admission_count BETWEEN 0 AND 20),
    CHECK(running_count BETWEEN 0 AND 2 AND running_count<=admission_count),
    CHECK(revision>=1 AND pg_catalog.isfinite(updated_at))
);

CREATE TABLE analytics_dashboard_export_principal_counters (
    actor_id uuid PRIMARY KEY,
    admission_count bigint NOT NULL DEFAULT 0,
    revision bigint NOT NULL DEFAULT 1,
    updated_at timestamptz NOT NULL,
    CHECK(actor_id<>'00000000-0000-0000-0000-000000000000'::uuid),
    CHECK(admission_count BETWEEN 0 AND 5),
    CHECK(revision>=1 AND pg_catalog.isfinite(updated_at))
);

CREATE TABLE analytics_dashboard_export_capacity_reservations (
    export_id uuid PRIMARY KEY REFERENCES analytics_dashboard_exports(export_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    tenant_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    admission_held boolean NOT NULL,
    running_held boolean NOT NULL,
    running_generation bigint,
    acquired_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CHECK(tenant_id<>'00000000-0000-0000-0000-000000000000'::uuid
        AND actor_id<>'00000000-0000-0000-0000-000000000000'::uuid),
    CHECK((running_held AND admission_held AND running_generation BETWEEN 1 AND 3)
       OR (NOT running_held AND running_generation IS NULL)),
    CHECK(pg_catalog.isfinite(acquired_at) AND pg_catalog.isfinite(updated_at)
        AND updated_at>=acquired_at)
);

CREATE TABLE analytics_dashboard_export_capacity_receipts (
    export_id uuid NOT NULL REFERENCES analytics_dashboard_export_capacity_reservations(export_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    transition_kind text NOT NULL,
    attempt_generation bigint NOT NULL,
    tenant_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    occurred_at timestamptz NOT NULL,
    PRIMARY KEY(export_id,transition_kind,attempt_generation),
    CHECK(transition_kind IN (
        'admission_acquired','running_acquired','running_released','admission_released'
    )),
    CHECK(attempt_generation BETWEEN 0 AND 3),
    CHECK(pg_catalog.isfinite(occurred_at))
);

INSERT INTO analytics_dashboard_export_capacity_counters(
    tenant_id,admission_count,running_count,updated_at
)
SELECT tenant_id,pg_catalog.count(*)::bigint,
       pg_catalog.count(*) FILTER (WHERE state='running')::bigint,
       pg_catalog.max(updated_at)
FROM analytics_dashboard_exports
WHERE state IN ('queued','running')
GROUP BY tenant_id;

INSERT INTO analytics_dashboard_export_principal_counters(
    actor_id,admission_count,updated_at
)
SELECT actor_id,pg_catalog.count(*)::bigint,pg_catalog.max(updated_at)
FROM analytics_dashboard_exports
WHERE state IN ('queued','running')
GROUP BY actor_id;

INSERT INTO analytics_dashboard_export_capacity_reservations(
    export_id,tenant_id,actor_id,admission_held,running_held,
    running_generation,acquired_at,updated_at
)
SELECT export_id,tenant_id,actor_id,true,state='running',
       CASE WHEN state='running' THEN attempt_generation END,
       created_at,updated_at
FROM analytics_dashboard_exports
WHERE state IN ('queued','running');

INSERT INTO analytics_dashboard_export_capacity_receipts(
    export_id,transition_kind,attempt_generation,tenant_id,actor_id,occurred_at
)
SELECT export_id,'admission_acquired',0,tenant_id,actor_id,created_at
FROM analytics_dashboard_exports
WHERE state IN ('queued','running');

INSERT INTO analytics_dashboard_export_capacity_receipts(
    export_id,transition_kind,attempt_generation,tenant_id,actor_id,occurred_at
)
SELECT export_id,'running_acquired',attempt_generation,tenant_id,actor_id,updated_at
FROM analytics_dashboard_exports
WHERE state='running';

-- +goose StatementBegin
CREATE FUNCTION analytics_guard_dashboard_export_capacity_receipt_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF TG_OP IN ('UPDATE','DELETE') THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='Analytics dashboard export capacity receipt is immutable';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER analytics_dashboard_export_capacity_receipts_guard
BEFORE UPDATE OR DELETE ON analytics_dashboard_export_capacity_receipts
FOR EACH ROW EXECUTE FUNCTION analytics_guard_dashboard_export_capacity_receipt_v1();
ALTER TABLE analytics_dashboard_export_capacity_receipts
    ENABLE ALWAYS TRIGGER analytics_dashboard_export_capacity_receipts_guard;
REVOKE ALL ON FUNCTION analytics_guard_dashboard_export_capacity_receipt_v1()
FROM PUBLIC;

