-- +goose Up
CREATE TABLE platform.billing_platform_license_refresh_requests (
    operation_id uuid PRIMARY KEY,
    license_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    idempotency_key text NOT NULL,
    request_sha256 text NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_platform_license_refresh_request_idempotency_uk
        UNIQUE (environment_id, actor_id, idempotency_key),
    CONSTRAINT billing_platform_license_refresh_request_head_fk
        FOREIGN KEY (license_id, environment_id)
        REFERENCES platform.billing_license_heads (license_id, environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_platform_license_refresh_request_environment_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.environment_identity (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_platform_license_refresh_request_tenant_fk
        FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_platform_license_refresh_request_shape CHECK (
        pg_catalog.uuid_extract_version(operation_id) = 7
        AND pg_catalog.uuid_extract_version(environment_id) = 7
        AND license_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND actor_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$'
        AND request_sha256 ~ '^[0-9a-f]{64}$'
        AND pg_catalog.isfinite(created_at)
    )
);

-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_platform_license_refresh_request_v1()
RETURNS trigger
LANGUAGE plpgsql SET search_path TO pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'billing platform license refresh request is immutable';
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER billing_platform_license_refresh_request_immutable
BEFORE UPDATE OR DELETE
ON platform.billing_platform_license_refresh_requests
FOR EACH ROW
EXECUTE FUNCTION platform.billing_guard_platform_license_refresh_request_v1();

REVOKE ALL ON FUNCTION
platform.billing_guard_platform_license_refresh_request_v1()
FROM PUBLIC;
