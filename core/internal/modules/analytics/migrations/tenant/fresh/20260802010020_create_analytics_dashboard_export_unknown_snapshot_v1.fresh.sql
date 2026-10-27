-- +goose Up
ALTER TABLE analytics_dashboard_export_platform_attempts
    ADD CONSTRAINT analytics_dashboard_export_unknown_identity_key UNIQUE (
        export_id,attempt_generation,job_id,platform_attempt_id,lease_owner,
        platform_effect_id,platform_effect_request_sha256
    );

CREATE TABLE analytics_dashboard_export_unknown_resolutions (
    tenant_id uuid NOT NULL,
    export_id uuid NOT NULL,
    attempt_generation bigint NOT NULL,
    job_id uuid NOT NULL,
    platform_attempt_id uuid NOT NULL,
    lease_owner uuid NOT NULL,
    platform_effect_id uuid NOT NULL,
    platform_effect_request_sha256 bytea NOT NULL,
    resolution_receipt_id uuid NOT NULL UNIQUE,
    resolution_kind text NOT NULL,
    resolution_sha256 bytea NOT NULL,
    object_key text NOT NULL,
    object_version text,
    reconcile_provider_request_id text NOT NULL,
    reconcile_evidence_sha256 bytea NOT NULL,
    upload_completed_at timestamptz,
    reconcile_verified_at timestamptz NOT NULL,
    delete_provider_request_id text,
    delete_response_sha256 bytea,
    absence_sha256 bytea NOT NULL,
    delete_verified_at timestamptz,
    terminal_state text NOT NULL,
    quota_released_bytes bigint NOT NULL,
    resolved_at timestamptz NOT NULL,
    PRIMARY KEY(export_id,attempt_generation),
    FOREIGN KEY (
        export_id,attempt_generation,job_id,platform_attempt_id,lease_owner,
        platform_effect_id,platform_effect_request_sha256
    ) REFERENCES analytics_dashboard_export_platform_attempts(
        export_id,attempt_generation,job_id,platform_attempt_id,lease_owner,
        platform_effect_id,platform_effect_request_sha256
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK(tenant_id<>'00000000-0000-0000-0000-000000000000'::uuid
        AND pg_catalog.uuid_extract_version(export_id)=7
        AND pg_catalog.uuid_extract_version(job_id)=7
        AND pg_catalog.uuid_extract_version(platform_attempt_id)=7
        AND pg_catalog.uuid_extract_version(lease_owner)=7
        AND pg_catalog.uuid_extract_version(platform_effect_id)=7
        AND pg_catalog.uuid_extract_version(resolution_receipt_id)=7),
    CHECK(pg_catalog.octet_length(platform_effect_request_sha256)=32
        AND pg_catalog.encode(platform_effect_request_sha256,'hex')<>pg_catalog.repeat('0',64)
        AND pg_catalog.octet_length(resolution_sha256)=32
        AND pg_catalog.encode(resolution_sha256,'hex')<>pg_catalog.repeat('0',64)
        AND pg_catalog.octet_length(reconcile_evidence_sha256)=32
        AND pg_catalog.encode(reconcile_evidence_sha256,'hex')<>pg_catalog.repeat('0',64)
        AND pg_catalog.octet_length(absence_sha256)=32
        AND pg_catalog.encode(absence_sha256,'hex')<>pg_catalog.repeat('0',64)),
    CHECK(resolution_kind IN ('deleted','absent')),
    CHECK(pg_catalog.octet_length(object_key) BETWEEN 1 AND 1024
        AND pg_catalog.octet_length(reconcile_provider_request_id) BETWEEN 1 AND 512),
    CHECK((resolution_kind='absent' AND object_version IS NULL
            AND upload_completed_at IS NULL AND delete_provider_request_id IS NULL
            AND delete_response_sha256 IS NULL AND delete_verified_at IS NULL
            AND absence_sha256=reconcile_evidence_sha256) OR
          (resolution_kind='deleted' AND
            pg_catalog.octet_length(object_version) BETWEEN 1 AND 512
            AND pg_catalog.octet_length(delete_provider_request_id) BETWEEN 1 AND 512
            AND pg_catalog.octet_length(delete_response_sha256)=32
            AND pg_catalog.encode(delete_response_sha256,'hex')<>pg_catalog.repeat('0',64)
            AND upload_completed_at IS NOT NULL AND delete_verified_at IS NOT NULL)),
    CHECK(terminal_state IN ('failed','cancelled','expired')),
    CHECK(quota_released_bytes=2147483648),
    CHECK(pg_catalog.isfinite(resolved_at)
        AND pg_catalog.isfinite(reconcile_verified_at)
        AND reconcile_verified_at<=resolved_at
        AND (upload_completed_at IS NULL OR
            (pg_catalog.isfinite(upload_completed_at)
             AND upload_completed_at<=reconcile_verified_at))
        AND (delete_verified_at IS NULL OR
            (pg_catalog.isfinite(delete_verified_at)
             AND reconcile_verified_at<=delete_verified_at
             AND delete_verified_at<=resolved_at)))
);

-- +goose StatementBegin
CREATE FUNCTION analytics_guard_dashboard_export_unknown_resolution_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF TG_OP IN ('UPDATE','DELETE') THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='Analytics dashboard export unknown resolution is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE TRIGGER analytics_dashboard_export_unknown_resolutions_guard
BEFORE UPDATE OR DELETE ON analytics_dashboard_export_unknown_resolutions
FOR EACH ROW EXECUTE FUNCTION
    analytics_guard_dashboard_export_unknown_resolution_v1();
ALTER TABLE analytics_dashboard_export_unknown_resolutions
    ENABLE ALWAYS TRIGGER analytics_dashboard_export_unknown_resolutions_guard;

CREATE FUNCTION analytics_dashboard_export_unknown_snapshot_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_attempt_generation bigint,p_lease_owner uuid,
    p_effect_id uuid,p_effect_request_sha256 bytea
)
RETURNS TABLE(
    format text,object_key text,object_version text,key_id text,
    wrapped_dek bytea,wrap_nonce bytea,header_sha256 bytea,
    ciphertext_sha256 bytea,plaintext_sha256 bytea,ciphertext_bytes bigint,
    plaintext_bytes bigint,row_count bigint,content_type text,filename text,
    provider_request_id text,effect_started_at timestamptz,
    cancellation_requested boolean,expires_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_tenant_id uuid;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_lease_owner) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.octet_length(p_effect_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.encode(p_effect_request_sha256,'hex')=pg_catalog.repeat('0',64) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export unknown snapshot scope is invalid';
    END IF;
    RETURN QUERY
    SELECT item.format,receipt.object_key,item.object_version,receipt.key_id,
           receipt.wrapped_dek,receipt.wrap_nonce,receipt.header_sha256,
           receipt.ciphertext_sha256,receipt.plaintext_sha256,
           receipt.ciphertext_bytes,receipt.plaintext_bytes,receipt.row_count,
           item.content_type,item.filename,item.provider_request_id,
           item.effect_started_at,item.cancellation_requested_at IS NOT NULL,
           item.expires_at
    FROM analytics_dashboard_exports AS item
    JOIN analytics_dashboard_export_platform_attempts AS attempt
      ON attempt.export_id=item.export_id
     AND attempt.attempt_generation=item.attempt_generation
    JOIN analytics_dashboard_export_attempt_receipts AS receipt
      ON receipt.export_id=attempt.export_id
     AND receipt.attempt_generation=attempt.attempt_generation
    WHERE item.tenant_id=p_tenant_id AND item.export_id=p_export_id
      AND item.job_id=p_job_id AND item.state='running'
      AND item.attempt_generation=p_attempt_generation
      AND item.quota_state='held' AND item.platform_attempt_id=p_attempt_id
      AND item.platform_effect_id=p_effect_id
      AND item.platform_effect_request_sha256=p_effect_request_sha256
      AND item.effect_started_at IS NOT NULL
      AND item.lease_owner IS NULL
      AND item.failure_code='ARTIFACT_OUTCOME_UNKNOWN'
      AND attempt.attempt_generation=p_attempt_generation
      AND attempt.job_id=p_job_id AND attempt.platform_attempt_id=p_attempt_id
      AND attempt.lease_owner=p_lease_owner AND attempt.outcome='unknown'
      AND attempt.platform_effect_id=p_effect_id
      AND attempt.platform_effect_request_sha256=p_effect_request_sha256
      AND attempt.resolution_receipt_id IS NULL
      AND receipt.outcome='unknown'
      AND receipt.attempt_generation=p_attempt_generation
      AND receipt.platform_attempt_id=p_attempt_id
      AND receipt.platform_effect_id=p_effect_id
      AND receipt.platform_effect_request_sha256=p_effect_request_sha256
    FOR SHARE OF item,attempt,receipt;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export unknown snapshot differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    analytics_guard_dashboard_export_unknown_resolution_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_unknown_snapshot_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea
) FROM PUBLIC;
