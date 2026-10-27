-- +goose Up
CREATE TABLE system_resource_layout_installation_baselines (
    operation_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    user_revision_id uuid NOT NULL,
    organization_revision_id uuid NOT NULL,
    role_revision_id uuid NOT NULL,
    baseline_sha256 bytea NOT NULL,
    initialized_at timestamptz NOT NULL,
    CONSTRAINT system_resource_layout_installation_baselines_pkey
        PRIMARY KEY (operation_id),
    CONSTRAINT system_resource_layout_installation_baselines_tenant_key
        UNIQUE (tenant_id),
    CONSTRAINT system_resource_layout_installation_baselines_digest_check
        CHECK (pg_catalog.octet_length(baseline_sha256) = 32),
    CONSTRAINT system_resource_layout_installation_baselines_user_fk
        FOREIGN KEY (user_revision_id)
        REFERENCES system_resource_layout_revisions (id) ON DELETE RESTRICT,
    CONSTRAINT system_resource_layout_installation_baselines_organization_fk
        FOREIGN KEY (organization_revision_id)
        REFERENCES system_resource_layout_revisions (id) ON DELETE RESTRICT,
    CONSTRAINT system_resource_layout_installation_baselines_role_fk
        FOREIGN KEY (role_revision_id)
        REFERENCES system_resource_layout_revisions (id) ON DELETE RESTRICT
);

-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_installation_baseline_guard_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'system resource installation baseline is immutable';
END;
$function$;

CREATE FUNCTION system_resource_layout_installation_record_v1(
    p_operation_id uuid,
    p_tenant_id uuid,
    p_user_revision_id uuid,
    p_organization_revision_id uuid,
    p_role_revision_id uuid,
    p_now timestamptz
)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_existing system_resource_layout_installation_baselines%ROWTYPE;
    v_user_catalog bytea;
    v_user_presentation bytea;
    v_organization_catalog bytea;
    v_organization_presentation bytea;
    v_role_catalog bytea;
    v_role_presentation bytea;
    v_digest bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR p_now IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'system resource installation baseline denied';
    END IF;
    SELECT publication.catalog_sha256, publication.presentation_sha256
    INTO v_user_catalog, v_user_presentation
    FROM system_resource_layout_publications AS publication
    WHERE publication.resource_kind = 'user'
      AND publication.revision_id = p_user_revision_id
      AND publication.catalog_sha256 IS NOT NULL
      AND publication.presentation_sha256 IS NOT NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'user installation layout is not published';
    END IF;
    SELECT publication.catalog_sha256, publication.presentation_sha256
    INTO v_organization_catalog, v_organization_presentation
    FROM system_resource_layout_publications AS publication
    WHERE publication.resource_kind = 'organization'
      AND publication.revision_id = p_organization_revision_id
      AND publication.catalog_sha256 IS NOT NULL
      AND publication.presentation_sha256 IS NOT NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'organization installation layout is not published';
    END IF;
    SELECT publication.catalog_sha256, publication.presentation_sha256
    INTO v_role_catalog, v_role_presentation
    FROM system_resource_layout_publications AS publication
    WHERE publication.resource_kind = 'role'
      AND publication.revision_id = p_role_revision_id
      AND publication.catalog_sha256 IS NOT NULL
      AND publication.presentation_sha256 IS NOT NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'role installation layout is not published';
    END IF;
    v_digest := pg_catalog.sha256(
        pg_catalog.uuid_send(p_operation_id) || pg_catalog.uuid_send(p_tenant_id) ||
        pg_catalog.uuid_send(p_user_revision_id) || v_user_catalog || v_user_presentation ||
        pg_catalog.uuid_send(p_organization_revision_id) || v_organization_catalog ||
        v_organization_presentation || pg_catalog.uuid_send(p_role_revision_id) ||
        v_role_catalog || v_role_presentation
    );
    SELECT * INTO v_existing
    FROM system_resource_layout_installation_baselines AS baseline
    WHERE baseline.operation_id = p_operation_id
    FOR UPDATE;
    IF FOUND THEN
        IF v_existing.tenant_id IS DISTINCT FROM p_tenant_id
           OR v_existing.user_revision_id IS DISTINCT FROM p_user_revision_id
           OR v_existing.organization_revision_id IS DISTINCT FROM p_organization_revision_id
           OR v_existing.role_revision_id IS DISTINCT FROM p_role_revision_id
           OR v_existing.baseline_sha256 IS DISTINCT FROM v_digest THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'system resource installation replay conflicts';
        END IF;
        RETURN v_digest;
    END IF;
    INSERT INTO system_resource_layout_installation_baselines (
        operation_id, tenant_id, user_revision_id, organization_revision_id,
        role_revision_id, baseline_sha256, initialized_at
    ) VALUES (
        p_operation_id, p_tenant_id, p_user_revision_id, p_organization_revision_id,
        p_role_revision_id, v_digest, p_now
    );
    RETURN v_digest;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_installation_baseline_guard_v1()
    FROM PUBLIC;
REVOKE ALL ON FUNCTION system_resource_layout_installation_record_v1(
    uuid,uuid,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_installation_record_v1(
    uuid,uuid,uuid,uuid,uuid,timestamptz
) TO forgecrud_runtime;

CREATE TRIGGER system_resource_layout_installation_baselines_write_guard
BEFORE UPDATE OR DELETE ON system_resource_layout_installation_baselines
FOR EACH ROW EXECUTE FUNCTION system_resource_layout_installation_baseline_guard_v1();
ALTER TABLE system_resource_layout_installation_baselines
    ENABLE ALWAYS TRIGGER system_resource_layout_installation_baselines_write_guard;

CREATE TRIGGER system_resource_layout_installation_baselines_truncate_guard
BEFORE TRUNCATE ON system_resource_layout_installation_baselines
FOR EACH STATEMENT EXECUTE FUNCTION system_resource_layout_installation_baseline_guard_v1();
ALTER TABLE system_resource_layout_installation_baselines
    ENABLE ALWAYS TRIGGER system_resource_layout_installation_baselines_truncate_guard;
