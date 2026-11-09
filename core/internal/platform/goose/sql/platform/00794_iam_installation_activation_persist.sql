-- +goose Up
CREATE TABLE IF NOT EXISTS platform.iam_security_epochs (
    environment_id uuid PRIMARY KEY REFERENCES platform.tenancy_deployment_policy (environment_id) ON DELETE RESTRICT,
    tenant_id uuid NOT NULL,
    operation_id uuid NOT NULL,
    schema_name name NOT NULL,
    security_epoch bytea NOT NULL,
    generation bigint NOT NULL,
    rotated_at timestamptz NOT NULL,
    CONSTRAINT iam_security_epochs_epoch_length CHECK (octet_length(security_epoch) = 32),
    CONSTRAINT iam_security_epochs_generation_check CHECK (generation > 0)
);

CREATE TABLE IF NOT EXISTS platform.iam_login_directory (
    tenant_id uuid PRIMARY KEY REFERENCES platform.tenants (id) ON DELETE RESTRICT,
    operation_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    schema_name name NOT NULL,
    user_id uuid NOT NULL,
    identifier_id uuid NOT NULL,
    key_id text NOT NULL,
    identifier_hmac bytea NOT NULL,
    source_sha256 bytea NOT NULL,
    revision bigint NOT NULL,
    user_revision bigint NOT NULL,
    identifier_revision bigint NOT NULL,
    source_revision bigint NOT NULL,
    CONSTRAINT iam_login_directory_revision_check CHECK (revision = 1),
    CONSTRAINT iam_login_directory_hmac_length CHECK (octet_length(identifier_hmac) = 32),
    CONSTRAINT iam_login_directory_source_length CHECK (octet_length(source_sha256) = 32)
);

REVOKE ALL ON TABLE platform.iam_security_epochs FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON TABLE platform.iam_login_directory FROM PUBLIC, harpy_api, harpy_worker;

DROP FUNCTION platform.iam_installation_activation_assert_v2(uuid, uuid, uuid, name);
DROP FUNCTION platform.iam_installation_security_epoch_initialize_v2(uuid, uuid, uuid, name, timestamptz);
DROP FUNCTION platform.iam_installation_identity_directory_upsert_v1(uuid, uuid, uuid, name, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea);

-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_activation_assert_v2(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name)
RETURNS boolean LANGUAGE sql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
    SELECT platform.tenancy_activation_assert_v2(in_operation_id, in_tenant_id, in_schema_name)
       AND EXISTS (
            SELECT 1 FROM platform.tenancy_activation_facts AS fact
            WHERE fact.operation_id = in_operation_id AND fact.environment_id = in_environment_id
              AND fact.tenant_id = in_tenant_id AND fact.schema_name = in_schema_name::text
       )
       AND (
            platform.tenancy_activation_step_context_assert_v1(in_operation_id, 'iam.initialize_tenant_identity', in_tenant_id, in_schema_name)
            OR EXISTS (
                SELECT 1 FROM platform.iam_login_directory AS directory
                WHERE directory.operation_id = in_operation_id AND directory.environment_id = in_environment_id
                  AND directory.tenant_id = in_tenant_id AND directory.schema_name = in_schema_name
            )
       );
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_security_epoch_initialize_v2(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name, in_observed_at timestamptz)
RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE epoch bytea; stored platform.iam_security_epochs%ROWTYPE;
BEGIN
    IF NOT platform.iam_installation_activation_assert_v2(in_operation_id, in_environment_id, in_tenant_id, in_schema_name) OR in_observed_at IS NULL THEN
        RETURN false;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.tenancy_deployment_policy WHERE environment_id = in_environment_id) THEN
        RETURN false;
    END IF;
    epoch := pg_catalog.sha256(pg_catalog.convert_to('harpy:iam:security-epoch:v1|' || in_environment_id::text || '|' || in_tenant_id::text || '|' || in_operation_id::text, 'UTF8'));
    INSERT INTO platform.iam_security_epochs (environment_id, tenant_id, operation_id, schema_name, security_epoch, generation, rotated_at)
    VALUES (in_environment_id, in_tenant_id, in_operation_id, in_schema_name, epoch, 1, in_observed_at)
    ON CONFLICT (environment_id) DO NOTHING;
    SELECT * INTO stored FROM platform.iam_security_epochs WHERE environment_id = in_environment_id;
    IF NOT FOUND OR stored.tenant_id <> in_tenant_id OR stored.operation_id <> in_operation_id OR stored.security_epoch <> epoch THEN
        RAISE EXCEPTION 'IAM security epoch belongs to another installation' USING ERRCODE = '42501';
    END IF;
    RETURN true;
END;
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_identity_directory_upsert_v1(
    in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name,
    in_user_revision bigint, in_identifier_revision bigint, in_user_id uuid, in_identifier_id uuid,
    in_key_id text, in_identifier_hmac bytea, in_source_revision bigint, in_source_sha256 bytea
) RETURNS bigint LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE stored platform.iam_login_directory%ROWTYPE;
BEGIN
    IF NOT platform.iam_installation_activation_assert_v2(in_operation_id, in_environment_id, in_tenant_id, in_schema_name)
        OR in_user_revision < 1 OR in_identifier_revision < 1 OR in_source_revision < 1 OR in_key_id IS NULL OR in_key_id = ''
        OR pg_catalog.octet_length(in_identifier_hmac) <> 32 OR pg_catalog.octet_length(in_source_sha256) <> 32 THEN
        RETURN 0;
    END IF;
    SELECT * INTO stored FROM platform.iam_login_directory WHERE tenant_id = in_tenant_id;
    IF FOUND THEN
        IF stored.operation_id = in_operation_id AND stored.environment_id = in_environment_id AND stored.user_id = in_user_id
           AND stored.identifier_id = in_identifier_id AND stored.identifier_hmac = in_identifier_hmac AND stored.source_sha256 = in_source_sha256 THEN
            RETURN stored.revision;
        END IF;
        RAISE EXCEPTION 'IAM login directory belongs to another installation identity' USING ERRCODE = '42501';
    END IF;
    INSERT INTO platform.iam_login_directory (
        tenant_id, operation_id, environment_id, schema_name, user_id, identifier_id, key_id,
        identifier_hmac, source_sha256, revision, user_revision, identifier_revision, source_revision
    ) VALUES (
        in_tenant_id, in_operation_id, in_environment_id, in_schema_name, in_user_id, in_identifier_id, in_key_id,
        in_identifier_hmac, in_source_sha256, 1, in_user_revision, in_identifier_revision, in_source_revision
    );
    SELECT revision INTO stored.revision FROM platform.iam_login_directory WHERE tenant_id = in_tenant_id AND operation_id = in_operation_id;
    RETURN stored.revision;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_installation_activation_assert_v2(uuid, uuid, uuid, name) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_installation_security_epoch_initialize_v2(uuid, uuid, uuid, name, timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_installation_identity_directory_upsert_v1(uuid, uuid, uuid, name, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_installation_activation_assert_v2(uuid, uuid, uuid, name) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.iam_installation_security_epoch_initialize_v2(uuid, uuid, uuid, name, timestamptz) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.iam_installation_identity_directory_upsert_v1(uuid, uuid, uuid, name, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_installation_identity_directory_upsert_v1(uuid, uuid, uuid, name, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea);
DROP FUNCTION platform.iam_installation_security_epoch_initialize_v2(uuid, uuid, uuid, name, timestamptz);
DROP FUNCTION platform.iam_installation_activation_assert_v2(uuid, uuid, uuid, name);
DROP TABLE IF EXISTS platform.iam_login_directory;
DROP TABLE IF EXISTS platform.iam_security_epochs;
-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_activation_assert_v2(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_security_epoch_initialize_v2(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name, in_observed_at timestamptz)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_identity_directory_upsert_v1(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name, in_bound_bigint_5 bigint, in_bound_bigint_6 bigint, in_user_id uuid, in_identifier_id uuid, in_key_id text, in_identifier_hmac bytea, in_bound_bigint_11 bigint, in_source_sha256 bytea)
RETURNS bigint LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT 0::bigint; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_installation_activation_assert_v2(uuid, uuid, uuid, name) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_installation_security_epoch_initialize_v2(uuid, uuid, uuid, name, timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_installation_identity_directory_upsert_v1(uuid, uuid, uuid, name, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_installation_activation_assert_v2(uuid, uuid, uuid, name) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.iam_installation_security_epoch_initialize_v2(uuid, uuid, uuid, name, timestamptz) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.iam_installation_identity_directory_upsert_v1(uuid, uuid, uuid, name, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea) TO harpy_api, harpy_worker;
