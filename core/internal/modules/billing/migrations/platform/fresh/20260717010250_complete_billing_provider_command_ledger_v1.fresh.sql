-- +goose Up
ALTER TABLE platform.billing_provider_commands
    ADD COLUMN actor_id uuid,
    ADD COLUMN active_job_id uuid,
    ADD COLUMN finalizer_job_id uuid;

ALTER TABLE platform.billing_provider_commands
    ADD CONSTRAINT billing_provider_commands_route_check CHECK (
        (actor_id IS NULL AND active_job_id IS NULL AND finalizer_job_id IS NULL)
        OR (actor_id IS NOT NULL
            AND actor_id <> '00000000-0000-0000-0000-000000000000'::uuid
            AND pg_catalog.uuid_extract_version(active_job_id) = 7
            AND pg_catalog.uuid_extract_version(finalizer_job_id) = 7
            AND active_job_id <> finalizer_job_id)
    );

CREATE UNIQUE INDEX billing_provider_commands_active_job_key
    ON platform.billing_provider_commands (tenant_id, active_job_id)
    WHERE active_job_id IS NOT NULL;

CREATE UNIQUE INDEX billing_provider_commands_finalizer_job_key
    ON platform.billing_provider_commands (tenant_id, finalizer_job_id)
    WHERE finalizer_job_id IS NOT NULL;
