-- +goose Up
CREATE TABLE IF NOT EXISTS platform.iam_initial_admin_setup_stages (
    operation_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    identifier_id uuid NOT NULL,
    email text NOT NULL,
    token_id bytea NOT NULL,
    selector bytea NOT NULL,
    hash_key_id text NOT NULL,
    secret_hmac bytea NOT NULL,
    binding_sha256 bytea NOT NULL,
    artifact_id uuid NOT NULL UNIQUE,
    artifact_reference text NOT NULL,
    key_id text NOT NULL,
    nonce bytea NOT NULL,
    ciphertext bytea NOT NULL,
    tag bytea NOT NULL,
    aad_sha256 bytea NOT NULL,
    ciphertext_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL
);

REVOKE ALL ON TABLE platform.iam_initial_admin_setup_stages FROM PUBLIC, harpy_api, harpy_worker;
GRANT SELECT ON TABLE platform.iam_initial_admin_setup_stages TO harpy_api, harpy_worker;

DROP FUNCTION IF EXISTS platform.iam_initial_admin_setup_stage_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea, timestamptz);
DROP FUNCTION IF EXISTS platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text);

-- +goose StatementBegin
CREATE FUNCTION platform.iam_initial_admin_setup_stage_v1(
    in_operation_id uuid, in_tenant_id uuid, in_user_id uuid, in_identifier_id uuid, in_email text, in_token_id bytea,
    in_selector bytea, in_hash_key_id text, in_secret_hmac bytea, in_binding_sha256 bytea, in_artifact_id uuid,
    in_reference text, in_key_id text, in_nonce bytea, in_ciphertext bytea, in_tag bytea, in_aad_sha256 bytea,
    in_ciphertext_sha256 bytea, in_created_at timestamptz
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE stored platform.iam_initial_admin_setup_stages%ROWTYPE;
BEGIN
    IF in_operation_id IS NULL OR in_artifact_id IS NULL OR in_ciphertext IS NULL OR pg_catalog.octet_length(in_ciphertext) < 1
        OR in_ciphertext_sha256 IS NULL OR in_created_at IS NULL OR in_email IS NULL THEN
        RAISE EXCEPTION 'initial admin setup stage is invalid' USING ERRCODE = '22023';
    END IF;
    SELECT * INTO stored FROM platform.iam_initial_admin_setup_stages WHERE operation_id = in_operation_id;
    IF FOUND THEN
        IF stored.tenant_id = in_tenant_id AND stored.artifact_id = in_artifact_id AND stored.ciphertext_sha256 = in_ciphertext_sha256
           AND stored.user_id = in_user_id AND stored.email = in_email THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION 'initial admin setup stage does not match the stored artifact' USING ERRCODE = '23514';
    END IF;
    INSERT INTO platform.iam_initial_admin_setup_stages (
        operation_id, tenant_id, user_id, identifier_id, email, token_id, selector, hash_key_id, secret_hmac, binding_sha256,
        artifact_id, artifact_reference, key_id, nonce, ciphertext, tag, aad_sha256, ciphertext_sha256, created_at
    ) VALUES (
        in_operation_id, in_tenant_id, in_user_id, in_identifier_id, in_email, in_token_id, in_selector, in_hash_key_id,
        in_secret_hmac, in_binding_sha256, in_artifact_id, in_reference, in_key_id, in_nonce, in_ciphertext, in_tag,
        in_aad_sha256, in_ciphertext_sha256, in_created_at
    );
    RETURN false;
END;
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(
    in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_artifact_id uuid, in_artifact_reference text, in_schema_name name, in_manifest_sha256 text
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE; payload_sha text;
BEGIN
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations WHERE operation_id = in_operation_id;
    IF NOT FOUND OR operation.tenant_id <> in_tenant_id OR operation.environment_id IS DISTINCT FROM in_environment_id
        OR operation.schema_name IS DISTINCT FROM in_schema_name::text OR operation.manifest_sha256 IS DISTINCT FROM in_manifest_sha256 THEN
        RAISE EXCEPTION 'initial admin setup delivery does not match the activation' USING ERRCODE = '42501';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.iam_initial_admin_setup_stages WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id
        AND artifact_id = in_artifact_id AND artifact_reference = in_artifact_reference) THEN
        RAISE EXCEPTION 'initial admin setup delivery has no staged credential' USING ERRCODE = '23514';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.platform_async_installation_activation_receipts WHERE operation_id = in_operation_id AND receipt_kind = 'initial_admin_setup' AND artifact_id = in_artifact_id) THEN
        RETURN true;
    END IF;
    IF NOT platform.tenancy_activation_step_context_assert_v1(in_operation_id, 'activation.complete', in_tenant_id, in_schema_name) THEN
        RAISE EXCEPTION 'initial admin setup delivery is outside its activation transaction' USING ERRCODE = '42501';
    END IF;
    payload_sha := platform.tenancy_activation_digest_hex_v1('initial-admin-setup|' || in_operation_id::text || '|' || in_artifact_id::text);
    INSERT INTO platform.jobs (
        id, tenant_id, scope_kind, scope_id, authority_group_id, authority_group_generation, authority_domain, cutover_generation,
        writer_origin, kind, payload_revision, payload_ref, payload_sha256, dedupe_key, effect_class, priority, run_at, state, attempts, max_attempts, lease_generation
    ) VALUES (
        platform.tenancy_activation_digest_uuid_v1('async.admin.job|' || in_artifact_id::text), in_tenant_id, 'tenant', in_tenant_id,
        in_operation_id::text, 0, 'installation', 0, 'installation_activation', 'iam.initial_admin_setup.deliver', 1, in_artifact_reference,
        payload_sha, 'initial-admin:' || in_artifact_id::text, 'installation', 0, pg_catalog.clock_timestamp(), 'queued', 0, 8, 0
    ) ON CONFLICT (id) DO NOTHING;
    INSERT INTO platform.platform_async_installation_activation_receipts (
        operation_id, receipt_kind, environment_id, tenant_id, event_id, artifact_id, manifest_sha256, created_at
    ) VALUES (
        in_operation_id, 'initial_admin_setup', in_environment_id, in_tenant_id,
        platform.tenancy_activation_digest_uuid_v1('async.admin.event|' || in_artifact_id::text), in_artifact_id, in_manifest_sha256, pg_catalog.clock_timestamp()
    );
    RETURN false;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_initial_admin_setup_stage_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea, timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_initial_admin_setup_stage_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea, timestamptz) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text);
DROP FUNCTION platform.iam_initial_admin_setup_stage_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea, timestamptz);
DROP TABLE IF EXISTS platform.iam_initial_admin_setup_stages;
-- +goose StatementBegin
CREATE FUNCTION platform.iam_initial_admin_setup_stage_v1(in_operation_id uuid, in_tenant_id uuid, in_user_id uuid, in_identifier_id uuid, in_email text, in_id bytea, in_selector bytea, in_hash_key_id text, in_secret_hmac bytea, in_binding_sha256 bytea, in_artifact_id uuid, in_reference text, in_key_id text, in_nonce bytea, in_ciphertext bytea, in_tag bytea, in_aadsha256 bytea, in_ciphertext_sha256 bytea, in_created_at timestamptz)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_artifact_id uuid, in_artifact_reference text, in_schema_name name, in_manifest_sha256 text)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_initial_admin_setup_stage_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea, timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_initial_admin_setup_stage_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea, timestamptz) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text) TO harpy_api, harpy_worker;
