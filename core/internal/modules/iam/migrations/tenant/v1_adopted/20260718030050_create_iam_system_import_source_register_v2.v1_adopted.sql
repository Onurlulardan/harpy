-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_import_source_register_v2(
    p_scope_schema name,p_source_id uuid,p_receipt_id uuid,p_actor_id uuid,
    p_resource_kind text,p_token_version bigint,p_authorization_version bigint,
    p_idempotency_sha256 bytea,p_format text,p_object_key text,
    p_object_version text,p_header_sha256 bytea,p_ciphertext_sha256 bytea,
    p_ciphertext_bytes bigint,p_plaintext_sha256 bytea,p_plaintext_bytes bigint,
    p_key_id text,p_wrapped_dek bytea,p_wrap_nonce bytea,
    p_provider_request_id text,p_uploaded_at timestamptz
)
RETURNS TABLE(
    source_id uuid,resource_kind text,format text,plaintext_bytes bigint,
    plaintext_sha256 bytea,uploaded_at timestamptz,replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_existing iam_system_import_sources%ROWTYPE;
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'import',
        p_token_version,p_authorization_version,true
    ) THEN NULL; END IF;
    IF pg_catalog.uuid_extract_version(p_source_id)<>7 OR
       pg_catalog.uuid_extract_version(p_receipt_id)<>7 OR
       pg_catalog.octet_length(p_idempotency_sha256)<>32 OR
       p_format NOT IN ('csv','jsonl') OR
       p_object_key IS DISTINCT FROM
         'iam-system-import-sources/v1/'||p_source_id::text||'/1.fcart' OR
       pg_catalog.octet_length(p_object_version) NOT BETWEEN 1 AND 512 OR
       pg_catalog.octet_length(p_header_sha256)<>32 OR
       pg_catalog.octet_length(p_ciphertext_sha256)<>32 OR
       p_ciphertext_bytes NOT BETWEEN 1 AND 538968064 OR
       pg_catalog.octet_length(p_plaintext_sha256)<>32 OR
       p_plaintext_bytes NOT BETWEEN 1 AND 536870912 OR
       pg_catalog.octet_length(p_key_id) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_wrapped_dek) NOT BETWEEN 32 AND 1024 OR
       pg_catalog.octet_length(p_wrap_nonce)<>12 OR
       pg_catalog.octet_length(p_provider_request_id) NOT BETWEEN 1 AND 256 OR
       NOT pg_catalog.isfinite(p_uploaded_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM import source evidence is invalid';
    END IF;
    SELECT identity.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema();
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'iam-transfer-source:'||v_tenant_id::text||':'||p_actor_id::text,6103
        )
    );
    SELECT source.* INTO v_existing
    FROM iam_system_import_sources AS source
    WHERE source.actor_id=p_actor_id
      AND source.resource_kind=p_resource_kind
      AND source.idempotency_key_sha256=p_idempotency_sha256;
    IF FOUND THEN
        IF v_existing.state='deleted' OR
           v_existing.delete_after<=pg_catalog.transaction_timestamp() THEN
            RAISE EXCEPTION USING ERRCODE='P6103',
                MESSAGE='IAM import source idempotency receipt expired';
        END IF;
        IF v_existing.tenant_id IS DISTINCT FROM v_tenant_id OR
           v_existing.format IS DISTINCT FROM p_format OR
           v_existing.plaintext_sha256 IS DISTINCT FROM p_plaintext_sha256 OR
           v_existing.plaintext_bytes IS DISTINCT FROM p_plaintext_bytes THEN
            RAISE EXCEPTION USING ERRCODE='P6102',
                MESSAGE='IAM import source idempotency conflicts';
        END IF;
        RETURN QUERY SELECT v_existing.source_id,v_existing.resource_kind,
            v_existing.format,v_existing.plaintext_bytes,
            v_existing.plaintext_sha256,v_existing.uploaded_at,true;
        RETURN;
    END IF;
    INSERT INTO iam_system_import_sources(
        source_id,tenant_id,idempotency_receipt_id,resource_kind,actor_id,
        idempotency_key_sha256,format,object_key,object_version,header_sha256,
        ciphertext_sha256,ciphertext_bytes,plaintext_sha256,plaintext_bytes,
        key_id,wrapped_dek,wrap_nonce,provider_request_id,uploaded_at,delete_after
    ) VALUES(
        p_source_id,v_tenant_id,p_receipt_id,p_resource_kind,p_actor_id,
        p_idempotency_sha256,p_format,p_object_key,p_object_version,p_header_sha256,
        p_ciphertext_sha256,p_ciphertext_bytes,p_plaintext_sha256,p_plaintext_bytes,
        p_key_id,p_wrapped_dek,p_wrap_nonce,p_provider_request_id,p_uploaded_at,
        p_uploaded_at+interval '24 hours'
    );
    RETURN QUERY SELECT p_source_id,p_resource_kind,p_format,p_plaintext_bytes,
        p_plaintext_sha256,p_uploaded_at,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_import_source_register_v2(
    name,uuid,uuid,uuid,text,bigint,bigint,bytea,text,text,text,bytea,bytea,
    bigint,bytea,bigint,text,bytea,bytea,text,timestamptz
) FROM PUBLIC;
