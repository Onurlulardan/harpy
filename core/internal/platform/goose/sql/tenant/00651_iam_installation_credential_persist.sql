-- +goose Up
CREATE TABLE IF NOT EXISTS iam_installation_credentials (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    credential_id uuid NOT NULL,
    credential_version bigint NOT NULL,
    salt bytea NOT NULL,
    tag bytea NOT NULL,
    credential_revision bigint NOT NULL,
    marker_revision bigint NOT NULL,
    CONSTRAINT iam_installation_credentials_salt_length CHECK (octet_length(salt) = 16),
    CONSTRAINT iam_installation_credentials_tag_length CHECK (octet_length(tag) = 32),
    CONSTRAINT iam_installation_credentials_revision_check CHECK (credential_revision = 1 AND marker_revision = 1)
);

REVOKE ALL ON TABLE iam_installation_credentials FROM PUBLIC, harpy_api, harpy_worker;
-- +goose StatementBegin
DO $grant$ BEGIN
    EXECUTE pg_catalog.format('GRANT USAGE ON SCHEMA %I TO harpy_owner', pg_catalog.current_schema());
    EXECUTE pg_catalog.format('GRANT SELECT ON TABLE %I.iam_installation_credentials TO harpy_owner', pg_catalog.current_schema());
END $grant$;
-- +goose StatementEnd

DROP FUNCTION iam_installation_credential_initialize_v2(name, uuid, uuid, uuid, uuid, uuid, bigint, bytea, bytea);

-- +goose StatementBegin
CREATE FUNCTION iam_installation_credential_initialize_v2(
    in_schema_name name, in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_user_id uuid,
    in_credential_id uuid, in_credential_version bigint, in_salt bytea, in_tag bytea
) RETURNS TABLE (operation_id uuid, tenant_id uuid, user_id uuid, credential_id uuid, credential_version bigint, credential_revision bigint, marker_revision bigint)
LANGUAGE plpgsql VOLATILE SECURITY DEFINER
AS $fn$
DECLARE tenant_schema name; identity_tenant uuid; matched boolean;
BEGIN
    tenant_schema := (pg_catalog.current_schemas(false))[2];
    IF tenant_schema IS DISTINCT FROM in_schema_name OR tenant_schema::text !~ '^tenant_[0-9a-f]{32}$'
        OR pg_catalog.octet_length(in_salt) <> 16 OR pg_catalog.octet_length(in_tag) <> 32 OR in_credential_version < 1 THEN
        RAISE EXCEPTION 'installation credential is not bound to this tenant' USING ERRCODE = '42501';
    END IF;
    EXECUTE pg_catalog.format('SELECT tenant_id FROM %I.tenancy_schema_identity WHERE singleton', tenant_schema) INTO identity_tenant;
    IF identity_tenant IS DISTINCT FROM in_tenant_id OR NOT platform.tenancy_activation_assert_v2(in_operation_id, in_tenant_id, tenant_schema) THEN
        RAISE EXCEPTION 'installation credential does not match the activation' USING ERRCODE = '42501';
    END IF;
    EXECUTE pg_catalog.format('SELECT true FROM %I.iam_installation_credentials WHERE operation_id = $1 AND tenant_id = $2 AND user_id = $3 AND credential_id = $4 AND credential_version = $5 AND salt = $6 AND tag = $7', tenant_schema)
        INTO matched USING in_operation_id, in_tenant_id, in_user_id, in_credential_id, in_credential_version, in_salt, in_tag;
    IF matched THEN
        RETURN QUERY EXECUTE pg_catalog.format('SELECT operation_id, tenant_id, user_id, credential_id, credential_version, credential_revision, marker_revision FROM %I.iam_installation_credentials WHERE operation_id = $1', tenant_schema) USING in_operation_id;
        RETURN;
    END IF;
    IF NOT platform.tenancy_activation_step_context_assert_v1(in_operation_id, 'iam.initialize_tenant_identity', in_tenant_id, tenant_schema) THEN
        RAISE EXCEPTION 'installation credential capability is missing' USING ERRCODE = '42501';
    END IF;
    EXECUTE pg_catalog.format('INSERT INTO %I.iam_installation_credentials (operation_id, environment_id, tenant_id, user_id, credential_id, credential_version, salt, tag, credential_revision, marker_revision) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,1,1)', tenant_schema)
        USING in_operation_id, in_environment_id, in_tenant_id, in_user_id, in_credential_id, in_credential_version, in_salt, in_tag;
    PERFORM platform.tenancy_activation_step_consume_v1(in_operation_id, 'iam.initialize_tenant_identity', in_tenant_id, tenant_schema);
    RETURN QUERY EXECUTE pg_catalog.format('SELECT operation_id, tenant_id, user_id, credential_id, credential_version, credential_revision, marker_revision FROM %I.iam_installation_credentials WHERE operation_id = $1', tenant_schema) USING in_operation_id;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_credential_initialize_v2(name, uuid, uuid, uuid, uuid, uuid, bigint, bytea, bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_installation_credential_initialize_v2(name, uuid, uuid, uuid, uuid, uuid, bigint, bytea, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_installation_credential_initialize_v2(name, uuid, uuid, uuid, uuid, uuid, bigint, bytea, bytea);
DROP TABLE IF EXISTS iam_installation_credentials;
-- +goose StatementBegin
CREATE FUNCTION iam_installation_credential_initialize_v2(in_schema_name name, in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_user_id uuid, in_credential_id uuid, in_credential_version bigint, in_salt bytea, in_tag bytea)
RETURNS TABLE (operation_id uuid, tenant_id uuid, user_id uuid, credential_id uuid, credential_version bigint, credential_revision bigint, marker_revision bigint)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$ SELECT in_operation_id, in_tenant_id, in_user_id, in_credential_id, in_credential_version, 0::bigint, 0::bigint; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_installation_credential_initialize_v2(name, uuid, uuid, uuid, uuid, uuid, bigint, bytea, bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_installation_credential_initialize_v2(name, uuid, uuid, uuid, uuid, uuid, bigint, bytea, bytea) TO harpy_api, harpy_worker;
