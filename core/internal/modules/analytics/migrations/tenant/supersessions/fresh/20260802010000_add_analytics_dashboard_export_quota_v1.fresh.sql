-- +goose Up
-- +goose StatementBegin
ALTER TABLE analytics_dashboard_exports
    ADD COLUMN tenant_id uuid,
    ADD COLUMN quota_state text NOT NULL DEFAULT 'none';

DROP TRIGGER analytics_dashboard_exports_guard ON analytics_dashboard_exports;

DO $migration$
DECLARE
    v_tenant_id uuid;
    v_used bigint;
    v_reserved bigint;
BEGIN
    SELECT identity_row.tenant_id INTO v_tenant_id
    FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton
      AND identity_row.schema_name=pg_catalog.current_schema();
    IF NOT FOUND THEN
        IF EXISTS (SELECT 1 FROM analytics_dashboard_exports) THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='analytics dashboard export quota backfill is missing tenant identity';
        END IF;
        RETURN;
    END IF;

    UPDATE analytics_dashboard_exports
    SET tenant_id=v_tenant_id,
        quota_state=CASE
            WHEN state='deleted' THEN 'released'
            WHEN state='ready' OR
                 (state='deleting' AND object_version IS NOT NULL) THEN 'used'
            WHEN state IN ('running','deleting') AND
                 effect_started_at IS NOT NULL THEN 'held'
            ELSE 'none'
        END;

    INSERT INTO platform.artifact_quota_counters(tenant_id)
    VALUES (v_tenant_id)
    ON CONFLICT (tenant_id) DO NOTHING;

    SELECT COALESCE(pg_catalog.sum(ciphertext_bytes)
               FILTER (WHERE quota_state='used'),0),
           2147483648 * pg_catalog.count(*)
               FILTER (WHERE quota_state='held')
    INTO v_used,v_reserved
    FROM analytics_dashboard_exports;

    IF v_used>0 OR v_reserved>0 THEN
        UPDATE platform.artifact_quota_counters
        SET used_ciphertext_bytes=used_ciphertext_bytes+v_used,
            reserved_ciphertext_bytes=reserved_ciphertext_bytes+v_reserved,
            revision=revision+1,
            updated_at=pg_catalog.clock_timestamp()
        WHERE tenant_id=v_tenant_id
          AND used_ciphertext_bytes+reserved_ciphertext_bytes+
              v_used+v_reserved<=21474836480;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6105',
                MESSAGE='analytics dashboard export quota backfill exceeds tenant limit';
        END IF;
    END IF;
END;
$migration$;

ALTER TABLE analytics_dashboard_exports
    ALTER COLUMN tenant_id SET NOT NULL,
    ADD CONSTRAINT analytics_dashboard_exports_tenant_fk
        FOREIGN KEY(tenant_id) REFERENCES platform.tenants(id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    ADD CONSTRAINT analytics_dashboard_exports_quota_check CHECK (
        quota_state IN ('none','held','used','released')
        AND (state<>'ready' OR quota_state='used')
        AND (state<>'deleted' OR quota_state='released')
    );

CREATE TRIGGER analytics_dashboard_exports_guard
BEFORE UPDATE OR DELETE ON analytics_dashboard_exports
FOR EACH ROW EXECUTE FUNCTION analytics_guard_dashboard_export_v1();
ALTER TABLE analytics_dashboard_exports
    ENABLE ALWAYS TRIGGER analytics_dashboard_exports_guard;
-- +goose StatementEnd
