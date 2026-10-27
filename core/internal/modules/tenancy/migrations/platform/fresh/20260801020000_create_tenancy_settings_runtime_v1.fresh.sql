-- +goose Up
ALTER TABLE platform.tenants
    ADD COLUMN settings_revision bigint NOT NULL DEFAULT 1,
    ADD CONSTRAINT tenancy_tenants_settings_revision_check
        CHECK (settings_revision > 0);

CREATE TABLE platform.tenancy_settings_operation_receipts (
    tenant_id uuid NOT NULL REFERENCES platform.tenants(id),
    idempotency_key text NOT NULL,
    request_sha256 bytea NOT NULL CHECK (pg_catalog.length(request_sha256) = 32),
    expected_revision bigint NOT NULL CHECK (expected_revision > 0),
    result_revision bigint NOT NULL CHECK (result_revision > 0),
    created_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, idempotency_key),
    CHECK (idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$')
);

CREATE TRIGGER tenancy_settings_operation_receipts_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenancy_settings_operation_receipts
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_mutation();

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_settings_read_v1(p_tenant_id uuid)
RETURNS TABLE (
    id uuid, name text, slug text, status text, settings_revision bigint,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR p_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='Tenant settings read denied';
    END IF;
    RETURN QUERY SELECT tenant.id, tenant.name::text, tenant.slug::text,
        tenant.status::text, tenant.settings_revision, tenant.created_at,
        tenant.updated_at
    FROM platform.tenants AS tenant
    WHERE tenant.id=p_tenant_id AND tenant.deleted_at IS NULL;
END;
$function$;
CREATE FUNCTION platform.tenancy_settings_update_v1(
    p_tenant_id uuid, p_name text, p_expected_revision bigint,
    p_idempotency_key text, p_request_sha256 bytea
) RETURNS TABLE (result_revision bigint, replay boolean, outcome text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_current bigint;
    v_digest bytea;
    v_expected bigint;
    v_result bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR p_tenant_id IS NULL OR
       p_expected_revision < 1 OR p_request_sha256 IS NULL OR
       pg_catalog.length(p_request_sha256) <> 32 OR p_idempotency_key IS NULL OR
       p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$' OR
       (p_name IS NOT NULL AND (p_name <> pg_catalog.btrim(p_name) OR
        pg_catalog.char_length(p_name) NOT BETWEEN 1 AND 255)) THEN
        RAISE EXCEPTION USING ERRCODE='22023', MESSAGE='Tenant settings update invalid';
    END IF;
    SELECT receipt.request_sha256, receipt.expected_revision, receipt.result_revision
      INTO v_digest, v_expected, v_result
    FROM platform.tenancy_settings_operation_receipts AS receipt
    WHERE receipt.tenant_id=p_tenant_id AND receipt.idempotency_key=p_idempotency_key;
    IF FOUND THEN
        RETURN QUERY SELECT v_result,
            v_digest=p_request_sha256 AND v_expected=p_expected_revision,
            CASE WHEN v_digest=p_request_sha256 AND v_expected=p_expected_revision
                 THEN 'replay' ELSE 'idempotency_conflict' END;
        RETURN;
    END IF;
    SELECT tenant.settings_revision INTO v_current
    FROM platform.tenants AS tenant
    WHERE tenant.id=p_tenant_id AND tenant.deleted_at IS NULL FOR UPDATE;
    IF NOT FOUND THEN
        RETURN QUERY SELECT 0::bigint, false, 'not_found'::text;
        RETURN;
    END IF;
    SELECT receipt.request_sha256, receipt.expected_revision, receipt.result_revision
      INTO v_digest, v_expected, v_result
    FROM platform.tenancy_settings_operation_receipts AS receipt
    WHERE receipt.tenant_id=p_tenant_id AND receipt.idempotency_key=p_idempotency_key;
    IF FOUND THEN
        RETURN QUERY SELECT v_result,
            v_digest=p_request_sha256 AND v_expected=p_expected_revision,
            CASE WHEN v_digest=p_request_sha256 AND v_expected=p_expected_revision
                 THEN 'replay' ELSE 'idempotency_conflict' END;
        RETURN;
    END IF;
    IF v_current <> p_expected_revision THEN
        RETURN QUERY SELECT v_current, false, 'revision_conflict'::text;
        RETURN;
    END IF;
    UPDATE platform.tenants AS tenant SET name=COALESCE(p_name,tenant.name),
        settings_revision=tenant.settings_revision+1,
        updated_at=pg_catalog.transaction_timestamp()
    WHERE tenant.id=p_tenant_id RETURNING tenant.settings_revision INTO v_result;
    INSERT INTO platform.tenancy_settings_operation_receipts(
        tenant_id,idempotency_key,request_sha256,expected_revision,
        result_revision,created_at
    ) VALUES (p_tenant_id,p_idempotency_key,p_request_sha256,
        p_expected_revision,v_result,pg_catalog.transaction_timestamp());
    RETURN QUERY SELECT v_result, false, 'applied'::text;
END;
$function$;
CREATE FUNCTION platform.tenancy_v1_settings_fence_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR NEW.id IS DISTINCT FROM OLD.id OR
       NEW.slug IS DISTINCT FROM OLD.slug OR
       NEW.created_at IS DISTINCT FROM OLD.created_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 tenant settings identity mutation denied';
    END IF;
    UPDATE platform.tenants AS tenant SET
        name=NEW.name, status=NEW.status, settings=NEW.settings,
        owner_id=NEW.owner_id, updated_at=NEW.updated_at,
        deleted_at=NEW.deleted_at,
        settings_revision=tenant.settings_revision+
            CASE WHEN NEW.name IS DISTINCT FROM OLD.name OR
                      NEW.settings IS DISTINCT FROM OLD.settings
                 THEN 1 ELSE 0 END
    WHERE tenant.id=OLD.id;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_v1_settings_fence_v1() FROM PUBLIC;

CREATE TRIGGER tenants_v1_settings_revision_fence
INSTEAD OF UPDATE ON public.tenants
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_v1_settings_fence_v1();

REVOKE ALL ON FUNCTION platform.tenancy_settings_read_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_settings_update_v1(uuid,text,bigint,text,bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_settings_read_v1(uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_settings_update_v1(uuid,text,bigint,text,bytea) TO harpy_runtime;
