-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_document_backfill_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_request_sha256 bytea,
    p_resource_kind text,
    p_resource_id uuid,
    p_expected_source_schema_id uuid,
    p_target_schema_id uuid,
    p_profile_document bytea,
    p_profile_sha256 bytea,
    p_expected_revision bigint,
    p_actor_id uuid,
    p_manage_permission_id uuid,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid,
    result_revision bigint,
    authorization_clock bigint,
    replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz;
    v_values jsonb;
    v_receipt_command text;
    v_receipt_request bytea;
    v_receipt_id uuid;
    v_receipt_revision bigint;
    v_receipt_clock bigint;
    v_clock bigint;
    v_revision bigint;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    IF p_operation_id IS NULL OR p_resource_id IS NULL OR
       p_expected_source_schema_id IS NULL OR p_target_schema_id IS NULL OR
       p_expected_source_schema_id IS NOT DISTINCT FROM p_target_schema_id OR
       p_resource_kind NOT IN ('user', 'organization', 'role') OR
       p_expected_revision < 1 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_profile_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_profile_document) > 1048576 OR
       pg_catalog.sha256(p_profile_document) IS DISTINCT FROM p_profile_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile document backfill input is invalid';
    END IF;
    v_values := pg_catalog.convert_from(p_profile_document, 'UTF8')::jsonb;
    IF pg_catalog.jsonb_typeof(v_values) <> 'object' OR EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_each(v_values) AS entry(key, value)
        WHERE entry.key !~
            '^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
           OR pg_catalog.jsonb_typeof(entry.value) = 'null'
           OR NOT EXISTS (
                SELECT 1 FROM iam_profile_field_revisions AS field
                WHERE field.schema_version_id = p_target_schema_id
                  AND field.resource_kind = p_resource_kind
                  AND field.field_id = entry.key::uuid
           )
    ) OR EXISTS (
        SELECT 1 FROM iam_profile_field_revisions AS required
        WHERE required.schema_version_id = p_target_schema_id
          AND required.resource_kind = p_resource_kind
          AND required.field_state = 'active' AND required.required
          AND NOT v_values ? required.field_id::text
    ) OR NOT EXISTS (
        SELECT 1 FROM iam_profile_schema_versions AS target
        WHERE target.id = p_target_schema_id
          AND target.resource_kind = p_resource_kind AND target.state = 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile document backfill contract is invalid';
    END IF;
    IF NOT iam_profile_actor_assert_v1(
        p_scope_schema, p_actor_id, p_manage_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile document backfill authority is denied';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_id,
           receipt.result_revision, receipt.clock_value
    INTO v_receipt_command, v_receipt_request, v_receipt_id,
         v_receipt_revision, v_receipt_clock
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id OPERATOR(pg_catalog.=) p_operation_id;
    IF FOUND THEN
        IF v_receipt_command IS DISTINCT FROM 'iam.profile_schema.backfill' OR
           v_receipt_request IS DISTINCT FROM p_request_sha256 OR
           v_receipt_id IS DISTINCT FROM p_resource_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM profile backfill operation identity was reused';
        END IF;
        RETURN QUERY SELECT v_receipt_id, v_receipt_revision,
            v_receipt_clock, true;
        RETURN;
    END IF;
    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1, revision = revision + 1, updated_at = v_now
    WHERE singleton RETURNING clock_value INTO v_clock;
    IF p_resource_kind = 'user' THEN
        UPDATE iam_users
        SET profile_schema_version_id = p_target_schema_id,
            profile_values = v_values, profile_sha256 = p_profile_sha256,
            revision = revision + 1, updated_at = v_now
        WHERE id = p_resource_id
          AND profile_schema_version_id = p_expected_source_schema_id
          AND revision = p_expected_revision
        RETURNING revision INTO v_revision;
    ELSIF p_resource_kind = 'organization' THEN
        UPDATE iam_organizations
        SET profile_schema_version_id = p_target_schema_id,
            profile_values = v_values, profile_sha256 = p_profile_sha256,
            revision = revision + 1, updated_at = v_now
        WHERE id = p_resource_id
          AND profile_schema_version_id = p_expected_source_schema_id
          AND revision = p_expected_revision
        RETURNING revision INTO v_revision;
    ELSE
        UPDATE iam_roles
        SET profile_schema_version_id = p_target_schema_id,
            profile_values = v_values, profile_sha256 = p_profile_sha256,
            revision = revision + 1, updated_at = v_now
        WHERE id = p_resource_id
          AND profile_schema_version_id = p_expected_source_schema_id
          AND revision = p_expected_revision
        RETURNING revision INTO v_revision;
    END IF;
    IF v_clock IS NULL OR v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile document backfill compare-and-swap failed';
    END IF;
    INSERT INTO iam_authorization_operation_receipts (
        operation_id, command_key, request_sha256, clock_value,
        result_code, result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, 'iam.profile_schema.backfill', p_request_sha256,
        v_clock, 'iam.profile_schema.backfilled', p_resource_id, v_revision,
        p_profile_sha256, v_now
    );
    RETURN QUERY SELECT p_resource_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_document_backfill_v1(
    name,uuid,bytea,text,uuid,uuid,uuid,bytea,bytea,bigint,
    uuid,uuid,bigint,bigint
) FROM PUBLIC;

