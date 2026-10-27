-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_publish_license_derived_offering_v1(
    p_operation_id uuid, p_environment_id uuid,
    p_license_id uuid, p_licensee_organization_id uuid,
    p_deployment_mode text, p_delivery_kind text, p_issuer text,
    p_sequence bigint, p_snapshot_sha256 text,
    p_offering_version_id uuid, p_revision bigint,
    p_offering_sha256 text, p_effective_at timestamptz,
    p_manifest_sha256 text, p_offering_document bytea
)
RETURNS TABLE (
    operation_id uuid, offering_version_id uuid,
    revision bigint, offering_sha256 text, replay boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_publication platform.billing_license_offering_publications%ROWTYPE;
    v_snapshot platform.billing_license_snapshots%ROWTYPE;
    v_authority_id uuid;
    v_offering_id uuid;
    v_document jsonb;
    v_manifest text;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    v_offering_id := pg_catalog.gen_random_uuid();
    v_manifest :=
        '3a0e120dd180d30f0e9c4ccb20c5338554c5f9e6db8349bff5335d813a421e99';
    IF pg_catalog.uuid_extract_version(p_operation_id) <> 7
       OR pg_catalog.uuid_extract_version(p_environment_id) <> 7
       OR pg_catalog.uuid_extract_version(p_offering_version_id) <> 7
       OR p_license_id IS NULL OR p_licensee_organization_id IS NULL
       OR p_deployment_mode <> 'onpremise'
       OR p_delivery_kind NOT IN ('connected', 'airgap', 'development')
       OR pg_catalog.octet_length(p_issuer) NOT BETWEEN 1 AND 256
       OR p_sequence < 1 OR p_revision < 1
       OR p_snapshot_sha256 !~ '^[0-9a-f]{64}$'
       OR p_offering_sha256 !~ '^[0-9a-f]{64}$'
       OR p_manifest_sha256 <> v_manifest
       OR p_offering_sha256 IS DISTINCT FROM pg_catalog.encode(
            pg_catalog.sha256(p_offering_document), 'hex')
       OR NOT pg_catalog.isfinite(p_effective_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing license offering publication is invalid';
    END IF;
    v_document := pg_catalog.convert_from(
        p_offering_document, 'UTF8'
    )::jsonb;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'billing-license-offering|' || p_operation_id::text, 0
    ));
    SELECT publication.* INTO v_publication
    FROM platform.billing_license_offering_publications AS publication
    WHERE publication.operation_id = p_operation_id;
    IF FOUND THEN
        IF v_publication.environment_id IS DISTINCT FROM p_environment_id
           OR v_publication.license_id IS DISTINCT FROM p_license_id
           OR v_publication.licensee_organization_id IS DISTINCT FROM
                p_licensee_organization_id
           OR v_publication.delivery_kind IS DISTINCT FROM p_delivery_kind
           OR v_publication.issuer IS DISTINCT FROM p_issuer
           OR v_publication.snapshot_sequence IS DISTINCT FROM p_sequence
           OR v_publication.snapshot_sha256 IS DISTINCT FROM p_snapshot_sha256
           OR v_publication.offering_version_id IS DISTINCT FROM
                p_offering_version_id
           OR v_publication.revision IS DISTINCT FROM p_revision
           OR v_publication.offering_sha256 IS DISTINCT FROM p_offering_sha256
           OR v_publication.offering_document IS DISTINCT FROM p_offering_document
           OR v_publication.effective_at IS DISTINCT FROM p_effective_at
           OR v_publication.manifest_sha256 IS DISTINCT FROM p_manifest_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'billing license offering operation was reused';
        END IF;
        RETURN QUERY SELECT p_operation_id, p_offering_version_id,
            p_revision, p_offering_sha256, true;
        RETURN;
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_license_id::text || '|' || p_environment_id::text, 0
    ));
    SELECT snapshot.* INTO v_snapshot
    FROM platform.billing_license_heads AS head
    JOIN platform.billing_license_snapshots AS snapshot
      ON snapshot.snapshot_id = head.snapshot_id
    WHERE head.license_id = p_license_id
      AND head.environment_id = p_environment_id
      AND head.sequence = p_sequence
      AND head.payload_sha256 = p_snapshot_sha256
      AND snapshot.licensee_organization_id = p_licensee_organization_id
      AND snapshot.delivery_kind = p_delivery_kind
      AND snapshot.issuer = p_issuer
    FOR SHARE OF head, snapshot;
    IF NOT FOUND OR NOT platform.billing_license_offering_document_valid_v1(
        v_snapshot.snapshot_id, p_offering_version_id,
        p_revision, p_effective_at, v_document
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'verified license offering snapshot is unavailable';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.billing_license_offering_publications AS publication
        WHERE publication.license_id = p_license_id
          AND publication.environment_id = p_environment_id
          AND publication.snapshot_sequence = p_sequence
          AND publication.snapshot_sha256 = p_snapshot_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'license snapshot already published another offering';
    END IF;
    SELECT authority.authority_id INTO STRICT v_authority_id
    FROM platform.billing_global_authorities AS authority
    WHERE authority.scope_id = p_environment_id
      AND authority.scope_kind = 'global'
      AND authority.authority_group_id = 'billing'
      AND authority.authority_domain = 'billing';
    PERFORM platform.billing_license_offering_insert_v1(
        v_authority_id, p_operation_id, v_offering_id,
        p_offering_version_id, p_revision, p_effective_at,
        v_snapshot.snapshot_id
    );
    INSERT INTO platform.billing_license_offering_publications (
        operation_id, authority_id, license_id, licensee_organization_id,
        environment_id, delivery_kind, issuer, snapshot_id,
        snapshot_sequence, snapshot_sha256, offering_id,
        offering_version_id, revision, offering_sha256,
        offering_document, effective_at, manifest_sha256, created_at
    ) VALUES (
        p_operation_id, v_authority_id, p_license_id,
        p_licensee_organization_id, p_environment_id, p_delivery_kind,
        p_issuer, v_snapshot.snapshot_id, p_sequence, p_snapshot_sha256,
        v_offering_id, p_offering_version_id, p_revision,
        p_offering_sha256, p_offering_document, p_effective_at,
        p_manifest_sha256, v_now
    );
    RETURN QUERY SELECT p_operation_id, p_offering_version_id,
        p_revision, p_offering_sha256, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_publish_license_derived_offering_v1(
    uuid, uuid, uuid, uuid, text, text, text, bigint, text, uuid,
    bigint, text, timestamptz, text, bytea
) FROM PUBLIC;
