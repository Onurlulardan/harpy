-- +goose Up
CREATE TABLE catalog_purge_approval_receipts_v2 (
    receipt_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL CHECK (tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    installation_id uuid NOT NULL,
    package_id uuid NOT NULL CHECK (package_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    operation_id uuid NOT NULL,
    requested_by uuid NOT NULL,
    approved_by uuid NOT NULL,
    backup_receipt_id uuid NOT NULL,
    backup_sha256 text NOT NULL CHECK (backup_sha256 ~ '^[0-9a-f]{64}$'),
    confirmation_slug text NOT NULL CHECK (confirmation_slug ~ '^[a-z][a-z0-9-]{1,62}$'),
    reason text NOT NULL CHECK (length(reason) BETWEEN 1 AND 1024),
    purge_idempotency_key text NOT NULL
        CHECK (purge_idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{7,127}$'),
    request_sha256 text NOT NULL CHECK (request_sha256 ~ '^[0-9a-f]{64}$'),
    approval_idempotency_key text NOT NULL
        CHECK (approval_idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{7,127}$'),
    step_up_authenticated_at timestamptz NOT NULL,
    approved_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    state text NOT NULL CHECK (state IN ('AVAILABLE','CONSUMED')),
    approval_sha256 text NOT NULL CHECK (approval_sha256 ~ '^[0-9a-f]{64}$'),
    consumed_operation_id uuid,
    consumed_authority_sha256 text CHECK (consumed_authority_sha256 ~ '^[0-9a-f]{64}$'),
    consumed_at timestamptz,
    UNIQUE (approved_by, approval_idempotency_key),
    FOREIGN KEY (installation_id) REFERENCES catalog_installations(installation_id)
        ON DELETE RESTRICT,
    CHECK (requested_by <> approved_by),
    CHECK (step_up_authenticated_at <= approved_at),
    CHECK (step_up_authenticated_at >= approved_at - interval '10 minutes'),
    CHECK (expires_at > approved_at),
    CHECK ((state='AVAILABLE' AND consumed_operation_id IS NULL
        AND consumed_authority_sha256 IS NULL AND consumed_at IS NULL)
        OR (state='CONSUMED' AND consumed_operation_id IS NOT NULL
        AND consumed_authority_sha256 IS NOT NULL AND consumed_at IS NOT NULL))
);
-- +goose StatementBegin
CREATE FUNCTION catalog_create_purge_approval_v2(
    p_scope_schema name, p_receipt_id uuid, p_tenant_id uuid,
    p_installation_id uuid, p_package_id uuid,
    p_operation_id uuid, p_requested_by uuid, p_approved_by uuid,
    p_backup_receipt_id uuid, p_backup_sha256 text, p_confirmation_slug text,
    p_reason text, p_purge_idempotency_key text, p_request_sha256 text,
    p_approval_idempotency_key text, p_step_up_at timestamptz,
    p_approved_at timestamptz, p_expires_at timestamptz, p_approval_sha256 text
) RETURNS SETOF catalog_purge_approval_receipts_v2
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
DECLARE
    v_installation catalog_installations%ROWTYPE;
    v_existing catalog_purge_approval_receipts_v2%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT EXISTS (SELECT 1 FROM tenancy_schema_identity i WHERE i.singleton
           AND i.schema_name=pg_catalog.current_schema() AND i.tenant_id=p_tenant_id) OR
       pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       p_tenant_id IS NULL OR p_tenant_id='00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_installation_id) IS DISTINCT FROM 7 OR
       p_package_id IS NULL OR p_package_id='00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_requested_by) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_approved_by) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_backup_receipt_id) IS DISTINCT FROM 7 OR
       p_backup_sha256 !~ '^[0-9a-f]{64}$' OR p_request_sha256 !~ '^[0-9a-f]{64}$' OR
       p_approval_sha256 !~ '^[0-9a-f]{64}$' OR
       p_purge_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{7,127}$' OR
       p_approval_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{7,127}$' OR
       NOT isfinite(p_step_up_at) OR NOT isfinite(p_approved_at) OR NOT isfinite(p_expires_at) OR
       date_trunc('microseconds',p_step_up_at)<>p_step_up_at OR
       date_trunc('microseconds',p_approved_at)<>p_approved_at OR
       date_trunc('microseconds',p_expires_at)<>p_expires_at THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='Catalog purge approval identifiers are invalid';
    END IF;
    SELECT * INTO v_existing FROM catalog_purge_approval_receipts_v2
    WHERE approved_by=p_approved_by AND approval_idempotency_key=p_approval_idempotency_key
    FOR SHARE;
    IF FOUND THEN
        IF v_existing.request_sha256<>p_request_sha256 OR
           v_existing.tenant_id<>p_tenant_id OR
           v_existing.installation_id<>p_installation_id OR
           v_existing.package_id<>p_package_id OR v_existing.operation_id<>p_operation_id OR
           v_existing.requested_by<>p_requested_by OR
           v_existing.backup_receipt_id<>p_backup_receipt_id OR
           v_existing.backup_sha256<>p_backup_sha256 OR
           v_existing.confirmation_slug<>p_confirmation_slug OR
           v_existing.reason<>p_reason OR
           v_existing.purge_idempotency_key<>p_purge_idempotency_key THEN
            RAISE EXCEPTION USING ERRCODE='23505', MESSAGE='Catalog purge approval replay differs';
        END IF;
        RETURN NEXT v_existing;
        RETURN;
    END IF;
    SELECT * INTO v_installation FROM catalog_installations
    WHERE installation_id=p_installation_id FOR SHARE;
    IF v_installation.installation_id IS NULL OR
       v_installation.package_id<>p_package_id OR v_installation.state<>'RETIRED' OR
       p_requested_by=p_approved_by OR p_step_up_at>p_approved_at OR
       p_step_up_at<p_approved_at-interval '10 minutes' OR
       p_expires_at<=p_approved_at THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='Catalog purge approval authority differs';
    END IF;
    INSERT INTO catalog_purge_approval_receipts_v2(
        receipt_id,tenant_id,installation_id,package_id,operation_id,requested_by,
        approved_by,backup_receipt_id,backup_sha256,confirmation_slug,reason,
        purge_idempotency_key,request_sha256,approval_idempotency_key,
        step_up_authenticated_at,approved_at,expires_at,state,approval_sha256
    ) VALUES (
        p_receipt_id,p_tenant_id,p_installation_id,p_package_id,p_operation_id,p_requested_by,
        p_approved_by,p_backup_receipt_id,p_backup_sha256,p_confirmation_slug,p_reason,
        p_purge_idempotency_key,p_request_sha256,p_approval_idempotency_key,
        p_step_up_at,p_approved_at,p_expires_at,'AVAILABLE',p_approval_sha256
    ) RETURNING * INTO v_existing;
    RETURN NEXT v_existing;
END;
$function$;
CREATE FUNCTION catalog_consume_purge_approval_v2(
    p_scope_schema name, p_receipt_id uuid, p_tenant_id uuid,
    p_package_id uuid, p_operation_id uuid,
    p_requested_by uuid, p_approved_by uuid, p_backup_receipt_id uuid,
    p_confirmation_slug text, p_reason text, p_purge_idempotency_key text,
    p_approval_sha256 text, p_authority_sha256 text, p_consumed_at timestamptz
) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
DECLARE v_receipt catalog_purge_approval_receipts_v2%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT EXISTS (SELECT 1 FROM tenancy_schema_identity i WHERE i.singleton
           AND i.schema_name=pg_catalog.current_schema() AND i.tenant_id=p_tenant_id) OR
       pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       p_approval_sha256 !~ '^[0-9a-f]{64}$' OR
       p_authority_sha256 !~ '^[0-9a-f]{64}$' OR NOT isfinite(p_consumed_at) OR
       date_trunc('microseconds',p_consumed_at)<>p_consumed_at THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='Catalog purge approval consume authority is invalid';
    END IF;
    SELECT * INTO v_receipt FROM catalog_purge_approval_receipts_v2
    WHERE receipt_id=p_receipt_id FOR UPDATE;
    IF v_receipt.receipt_id IS NULL OR v_receipt.tenant_id<>p_tenant_id OR
       v_receipt.package_id<>p_package_id OR v_receipt.operation_id<>p_operation_id OR
       v_receipt.requested_by<>p_requested_by OR v_receipt.approved_by<>p_approved_by OR
       v_receipt.backup_receipt_id<>p_backup_receipt_id OR
       v_receipt.confirmation_slug<>p_confirmation_slug OR v_receipt.reason<>p_reason OR
       v_receipt.purge_idempotency_key<>p_purge_idempotency_key OR
       v_receipt.approval_sha256<>p_approval_sha256 OR
       p_consumed_at<v_receipt.approved_at OR p_consumed_at>v_receipt.expires_at THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='Catalog purge approval consumption differs';
    END IF;
    IF v_receipt.state='CONSUMED' THEN
        IF v_receipt.consumed_operation_id<>p_operation_id OR
           v_receipt.consumed_authority_sha256<>p_authority_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='23505', MESSAGE='Catalog purge approval already consumed';
        END IF;
        RETURN v_receipt.approval_sha256;
    END IF;
    UPDATE catalog_purge_approval_receipts_v2 SET state='CONSUMED',
        consumed_operation_id=p_operation_id,consumed_authority_sha256=p_authority_sha256,
        consumed_at=p_consumed_at WHERE receipt_id=p_receipt_id;
    RETURN v_receipt.approval_sha256;
END;
$function$;
CREATE FUNCTION catalog_guard_purge_approval_v2()
RETURNS trigger LANGUAGE plpgsql SET search_path=pg_catalog AS $function$
BEGIN
    IF TG_OP='DELETE' THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='Catalog purge approval is immutable';
    END IF;
    IF (OLD.receipt_id,OLD.tenant_id,OLD.installation_id,OLD.package_id,OLD.operation_id,
        OLD.requested_by,OLD.approved_by,OLD.backup_receipt_id,OLD.backup_sha256,
        OLD.confirmation_slug,OLD.reason,OLD.purge_idempotency_key,OLD.request_sha256,
        OLD.approval_idempotency_key,OLD.step_up_authenticated_at,OLD.approved_at,
        OLD.expires_at,OLD.approval_sha256) IS DISTINCT FROM
       (NEW.receipt_id,NEW.tenant_id,NEW.installation_id,NEW.package_id,NEW.operation_id,
        NEW.requested_by,NEW.approved_by,NEW.backup_receipt_id,NEW.backup_sha256,
        NEW.confirmation_slug,NEW.reason,NEW.purge_idempotency_key,NEW.request_sha256,
        NEW.approval_idempotency_key,NEW.step_up_authenticated_at,NEW.approved_at,
        NEW.expires_at,NEW.approval_sha256) OR OLD.state<>'AVAILABLE' OR
       NEW.state<>'CONSUMED' OR NEW.consumed_operation_id<>OLD.operation_id OR
       NEW.consumed_authority_sha256 IS NULL OR NEW.consumed_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000', MESSAGE='Catalog purge approval is immutable';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd
CREATE TRIGGER catalog_guard_purge_approval_v2 BEFORE UPDATE OR DELETE
ON catalog_purge_approval_receipts_v2 FOR EACH ROW
EXECUTE FUNCTION catalog_guard_purge_approval_v2();
ALTER TABLE catalog_purge_approval_receipts_v2 ENABLE ALWAYS TRIGGER catalog_guard_purge_approval_v2;
REVOKE ALL ON FUNCTION catalog_create_purge_approval_v2(name,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,text,text,timestamptz,timestamptz,timestamptz,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION catalog_consume_purge_approval_v2(
    name,uuid,uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,text,timestamptz) FROM PUBLIC;
GRANT SELECT ON TABLE catalog_purge_approval_receipts_v2 TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION catalog_create_purge_approval_v2(name,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,text,text,timestamptz,timestamptz,timestamptz,text) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION catalog_consume_purge_approval_v2(
    name,uuid,uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,text,timestamptz) TO forgecrud_runtime;
