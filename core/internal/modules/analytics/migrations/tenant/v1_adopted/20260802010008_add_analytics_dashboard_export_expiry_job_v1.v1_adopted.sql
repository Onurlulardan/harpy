-- +goose Up
ALTER TABLE analytics_dashboard_exports
    ADD COLUMN expiry_job_id uuid,
    ADD CONSTRAINT analytics_dashboard_exports_expiry_job_unique
        UNIQUE(expiry_job_id),
    ADD CONSTRAINT analytics_dashboard_exports_expiry_job_check CHECK (
        expiry_job_id IS NULL OR
        pg_catalog.uuid_extract_version(expiry_job_id)=7
    );

