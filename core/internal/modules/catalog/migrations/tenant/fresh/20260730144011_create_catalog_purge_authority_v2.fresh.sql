-- +goose Up
CREATE TABLE catalog_purge_authorities_v2 (
    operation_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL
        CHECK (tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    package_id uuid NOT NULL
        CHECK (package_id <> '00000000-0000-0000-0000-000000000000'::uuid),
    package_slug text NOT NULL CHECK (package_slug ~ '^[a-z][a-z0-9-]{1,62}$'),
    actor_id uuid NOT NULL,
    requested_by uuid NOT NULL,
    approved_by uuid NOT NULL,
    approval_receipt_id uuid NOT NULL,
    approval_sha256 text NOT NULL CHECK (approval_sha256 ~ '^[0-9a-f]{64}$'),
    backup_receipt_id uuid NOT NULL,
    backup_state text NOT NULL CHECK (backup_state = 'READY'),
    backup_sha256 text NOT NULL CHECK (backup_sha256 ~ '^[0-9a-f]{64}$'),
    confirmation_slug text NOT NULL,
    reason text NOT NULL CHECK (length(reason) BETWEEN 1 AND 1024),
    idempotency_key text NOT NULL
        CHECK (idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{7,127}$'),
    reference_epoch bigint NOT NULL CHECK (reference_epoch > 0),
    step_up_authenticated_at timestamptz NOT NULL,
    approval_step_up_at timestamptz NOT NULL,
    approved_at timestamptz NOT NULL,
    approval_expires_at timestamptz NOT NULL,
    requested_at timestamptz NOT NULL,
    cooling_until timestamptz NOT NULL,
    authority_sha256 text NOT NULL CHECK (authority_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT catalog_purge_authority_operation_fk
        FOREIGN KEY (operation_id) REFERENCES catalog_operation_runs(operation_id)
        ON DELETE RESTRICT,
    CHECK (actor_id = requested_by AND requested_by <> approved_by),
    CHECK (confirmation_slug = package_slug),
    CHECK (step_up_authenticated_at <= requested_at),
    CHECK (step_up_authenticated_at >= requested_at - interval '10 minutes'),
    CHECK (approval_step_up_at <= approved_at),
    CHECK (approval_step_up_at >= approved_at - interval '10 minutes'),
    CHECK (approved_at <= requested_at AND requested_at <= approval_expires_at),
    CHECK (cooling_until >= requested_at + interval '24 hours'),
    CHECK (date_trunc('microseconds', step_up_authenticated_at) = step_up_authenticated_at),
    CHECK (date_trunc('microseconds', requested_at) = requested_at),
    CHECK (date_trunc('microseconds', cooling_until) = cooling_until)
);

-- +goose StatementBegin
CREATE FUNCTION catalog_record_purge_authority_v2(
    p_operation_id uuid, p_tenant_id uuid, p_package_id uuid, p_package_slug text,
    p_actor_id uuid, p_requested_by uuid, p_approved_by uuid,
    p_approval_receipt_id uuid, p_approval_sha256 text,
    p_backup_receipt_id uuid, p_backup_state text, p_backup_sha256 text,
    p_confirmation_slug text, p_reason text, p_idempotency_key text,
    p_reference_epoch bigint, p_step_up_at timestamptz, p_approval_step_up_at timestamptz,
    p_approved_at timestamptz, p_approval_expires_at timestamptz, p_requested_at timestamptz,
    p_cooling_until timestamptz, p_authority_sha256 text
) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
DECLARE
    v_operation catalog_operation_runs%ROWTYPE;
    v_installation catalog_installations%ROWTYPE;
    v_existing catalog_purge_authorities_v2%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       p_tenant_id IS NULL OR
       p_tenant_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_package_id IS NULL OR
       p_package_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_actor_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_requested_by) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_approved_by) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_approval_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_backup_receipt_id) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Catalog purge authority identifiers are invalid';
    END IF;
    SELECT * INTO v_operation FROM catalog_operation_runs
    WHERE operation_id=p_operation_id FOR UPDATE;
    SELECT * INTO v_installation FROM catalog_installations
    WHERE installation_id=v_operation.installation_id FOR SHARE;
    IF v_operation.operation_id IS NULL OR v_operation.kind<>'PURGE' OR
       v_operation.package_id<>p_package_id OR v_operation.actor_id<>p_actor_id OR
       v_operation.idempotency_key<>p_idempotency_key OR
       v_operation.plan->>'slug'<>p_package_slug OR
       v_installation.reference_epoch<>p_reference_epoch OR
       v_operation.started_at<p_requested_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog purge operation authority differs';
    END IF;
    SELECT * INTO v_existing FROM catalog_purge_authorities_v2
    WHERE operation_id=p_operation_id;
    IF FOUND THEN
        IF v_existing.tenant_id<>p_tenant_id OR v_existing.package_id<>p_package_id OR
           v_existing.actor_id<>p_actor_id OR v_existing.requested_by<>p_requested_by OR
           v_existing.approved_by<>p_approved_by OR
           v_existing.approval_receipt_id<>p_approval_receipt_id OR
           v_existing.approval_sha256<>p_approval_sha256 OR
           v_existing.backup_receipt_id<>p_backup_receipt_id OR
           v_existing.backup_sha256<>p_backup_sha256 OR
           v_existing.confirmation_slug<>p_confirmation_slug OR
           v_existing.reason<>p_reason OR v_existing.idempotency_key<>p_idempotency_key OR
           v_existing.reference_epoch<>p_reference_epoch OR
           v_existing.step_up_authenticated_at<>p_step_up_at OR
           v_existing.approval_step_up_at<>p_approval_step_up_at OR
           v_existing.approved_at<>p_approved_at OR
           v_existing.approval_expires_at<>p_approval_expires_at OR
           v_existing.requested_at<>p_requested_at OR
           v_existing.cooling_until<>p_cooling_until OR
           v_existing.authority_sha256<>p_authority_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Catalog purge authority differs';
        END IF;
        RETURN v_existing.authority_sha256;
    END IF;
    INSERT INTO catalog_purge_authorities_v2(
        operation_id,tenant_id,package_id,package_slug,actor_id,requested_by,
        approved_by,approval_receipt_id,approval_sha256,backup_receipt_id,
        backup_state,backup_sha256,confirmation_slug,reason,idempotency_key,
        reference_epoch,step_up_authenticated_at,approval_step_up_at,approved_at,
        approval_expires_at,requested_at,
        cooling_until,authority_sha256
    ) VALUES (
        p_operation_id,p_tenant_id,p_package_id,p_package_slug,p_actor_id,p_requested_by,
        p_approved_by,p_approval_receipt_id,p_approval_sha256,p_backup_receipt_id,
        p_backup_state,p_backup_sha256,p_confirmation_slug,p_reason,p_idempotency_key,
        p_reference_epoch,p_step_up_at,p_approval_step_up_at,p_approved_at,
        p_approval_expires_at,p_requested_at,
        p_cooling_until,p_authority_sha256
    );
    RETURN p_authority_sha256;
END;
$function$;

CREATE FUNCTION catalog_purge_authority_immutable_v2()
RETURNS trigger LANGUAGE plpgsql SET search_path=pg_catalog AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Catalog purge authority is immutable';
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER catalog_purge_authority_immutable_v2
BEFORE UPDATE OR DELETE ON catalog_purge_authorities_v2
FOR EACH STATEMENT EXECUTE FUNCTION catalog_purge_authority_immutable_v2();
ALTER TABLE catalog_purge_authorities_v2
    ENABLE ALWAYS TRIGGER catalog_purge_authority_immutable_v2;

REVOKE ALL ON FUNCTION catalog_record_purge_authority_v2(
    uuid,uuid,uuid,text,uuid,uuid,uuid,uuid,text,uuid,text,text,text,text,text,
    bigint,timestamptz,timestamptz,timestamptz,timestamptz,timestamptz,
    timestamptz,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION catalog_purge_authority_immutable_v2() FROM PUBLIC;
GRANT SELECT ON TABLE catalog_purge_authorities_v2 TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION catalog_record_purge_authority_v2(
    uuid,uuid,uuid,text,uuid,uuid,uuid,uuid,text,uuid,text,text,text,text,text,
    bigint,timestamptz,timestamptz,timestamptz,timestamptz,timestamptz,
    timestamptz,text) TO forgecrud_runtime;
