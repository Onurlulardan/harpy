-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_demo_record_handoff_v1(
    p_scope_schema name, p_owner_generation bigint,
    p_handoff_id uuid, p_provision_receipt_id uuid,
    p_component_id uuid, p_package_id uuid, p_release_id uuid,
    p_forms_run_id uuid, p_request_sha256 bytea,
    p_form_count integer, p_row_count integer,
    p_readback_sha256 bytea, p_completed_at timestamptz
)
RETURNS TABLE(
    handoff_id uuid, provision_receipt_id uuid, component_id uuid,
    package_id uuid, release_id uuid, forms_run_id uuid,
    request_sha256 bytea,
    form_count integer, row_count integer, readback_sha256 bytea,
    completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_provision catalog_demo_provisions%ROWTYPE;
    v_payload catalog_demo_payloads%ROWTYPE;
    v_existing catalog_demo_handoffs%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_handoff_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_provision_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_package_id IS NULL OR
       p_package_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_forms_run_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_form_count < 1 OR p_row_count < p_form_count OR
       pg_catalog.octet_length(p_readback_sha256) IS DISTINCT FROM 32 OR
       p_readback_sha256 = pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       pg_catalog.date_trunc('microseconds',p_completed_at) <> p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Catalog demo handoff input is invalid';
    END IF;
    PERFORM catalog_demo_assert_current_authority_v1(
        p_scope_schema,p_owner_generation
    );
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'catalog:demo-handoff:' || p_provision_receipt_id::text,6001
    ));
    SELECT * INTO v_existing
    FROM catalog_demo_handoffs AS handoff
    WHERE handoff.provision_receipt_id=p_provision_receipt_id;
    IF FOUND THEN
        IF ROW(v_existing.component_id,v_existing.package_id,
               v_existing.release_id,v_existing.forms_run_id,
               v_existing.request_sha256,
               v_existing.form_count,v_existing.row_count,
               v_existing.readback_sha256,v_existing.completed_at)
           IS DISTINCT FROM
           ROW(p_component_id,p_package_id,p_release_id,p_forms_run_id,
               p_request_sha256,
               p_form_count,p_row_count,p_readback_sha256,p_completed_at) THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Catalog demo handoff replay differs';
        END IF;
        RETURN QUERY SELECT v_existing.handoff_id,
            v_existing.provision_receipt_id,v_existing.component_id,
            v_existing.package_id,v_existing.release_id,
            v_existing.forms_run_id,
            v_existing.request_sha256,v_existing.form_count,
            v_existing.row_count,v_existing.readback_sha256,
            v_existing.completed_at;
        RETURN;
    END IF;
    SELECT * INTO STRICT v_provision
    FROM catalog_demo_provisions AS provision
    WHERE provision.receipt_id=p_provision_receipt_id
      AND provision.component_id=p_component_id
      AND provision.package_id=p_package_id
      AND provision.release_id=p_release_id
      AND provision.owner_generation=p_owner_generation
      AND provision.request_sha256=p_request_sha256
    FOR UPDATE;
    SELECT * INTO STRICT v_payload
    FROM catalog_demo_payloads AS payload
    WHERE payload.package_id=v_provision.package_id
      AND payload.package_version=v_provision.package_version
      AND payload.component_id=p_component_id
      AND payload.package_id=p_package_id
      AND payload.release_id=p_release_id
      AND payload.form_count=p_form_count
      AND payload.row_count=p_row_count
    FOR SHARE;
    IF v_provision.ready OR p_completed_at < v_provision.created_at THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='Catalog demo handoff provision state differs';
    END IF;
    INSERT INTO catalog_demo_handoffs(
        handoff_id,provision_receipt_id,component_id,package_id,
        package_version,release_id,forms_run_id,content_sha256,
        request_sha256,form_count,row_count,readback_sha256,completed_at
    ) VALUES (
        p_handoff_id,p_provision_receipt_id,p_component_id,p_package_id,
        v_provision.package_version,p_release_id,p_forms_run_id,
        v_provision.content_sha256,p_request_sha256,p_form_count,p_row_count,
        p_readback_sha256,p_completed_at
    ) RETURNING * INTO v_existing;
    RETURN QUERY SELECT v_existing.handoff_id,
        v_existing.provision_receipt_id,v_existing.component_id,
        v_existing.package_id,v_existing.release_id,
        v_existing.forms_run_id,
        v_existing.request_sha256,v_existing.form_count,
        v_existing.row_count,v_existing.readback_sha256,
        v_existing.completed_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_demo_record_handoff_v1(
    name,bigint,uuid,uuid,uuid,uuid,uuid,uuid,bytea,
    integer,integer,bytea,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION catalog_demo_record_handoff_v1(
    name,bigint,uuid,uuid,uuid,uuid,uuid,uuid,bytea,
    integer,integer,bytea,timestamptz
) TO forgecrud_runtime;
