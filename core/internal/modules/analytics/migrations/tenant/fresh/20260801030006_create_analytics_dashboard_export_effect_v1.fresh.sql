-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_effect_start_v1(
    p_scope_schema name, p_export_id uuid, p_lease_owner uuid,
    p_attempt_generation bigint, p_object_key text, p_key_id text,
    p_wrapped_dek bytea, p_wrap_nonce bytea, p_header_sha256 bytea,
    p_ciphertext_sha256 bytea, p_plaintext_sha256 bytea,
    p_ciphertext_bytes bigint, p_plaintext_bytes bigint, p_row_count bigint,
    p_content_type text, p_filename text, p_started_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_format text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       p_lease_owner IS NULL OR
       p_lease_owner='00000000-0000-0000-0000-000000000000'::uuid OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.octet_length(p_object_key) NOT BETWEEN 10 AND 512 OR
       pg_catalog.octet_length(p_key_id) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_wrapped_dek) IS DISTINCT FROM 48 OR
       pg_catalog.octet_length(p_wrap_nonce) IS DISTINCT FROM 12 OR
       pg_catalog.octet_length(p_header_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_ciphertext_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_plaintext_sha256) IS DISTINCT FROM 32 OR
       p_ciphertext_bytes<35 OR p_plaintext_bytes<1 OR
       p_plaintext_bytes>67108864 OR p_row_count<0 OR
       pg_catalog.octet_length(p_content_type) NOT BETWEEN 3 AND 128 OR
       pg_catalog.octet_length(p_filename) NOT BETWEEN 5 AND 255 OR
       p_started_at IS NULL OR NOT pg_catalog.isfinite(p_started_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export effect input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT item.format INTO v_format
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id
      AND item.state='running'
      AND item.attempt_generation=p_attempt_generation
      AND item.lease_owner=p_lease_owner
      AND item.lease_until>p_started_at
      AND item.effect_started_at IS NULL
      AND item.cancellation_requested_at IS NULL
      AND item.expires_at>p_started_at
    FOR UPDATE;
    IF NOT FOUND OR
       (v_format='csv' AND p_content_type<>'text/csv; charset=utf-8') OR
       (v_format='xlsx' AND p_content_type<>
        'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet') OR
       (v_format='pdf' AND p_content_type<>'application/pdf') THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export attempt changed';
    END IF;
    UPDATE analytics_dashboard_exports
    SET effect_started_at=p_started_at,object_key=p_object_key,
        key_id=p_key_id,wrapped_dek=p_wrapped_dek,wrap_nonce=p_wrap_nonce,
        header_sha256=p_header_sha256,
        ciphertext_sha256=p_ciphertext_sha256,
        plaintext_sha256=p_plaintext_sha256,
        ciphertext_bytes=p_ciphertext_bytes,
        plaintext_bytes=p_plaintext_bytes,row_count=p_row_count,
        content_type=p_content_type,filename=p_filename,
        row_revision=row_revision+1,updated_at=p_started_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
	INSERT INTO analytics_dashboard_export_attempt_receipts(
		export_id,attempt_generation,object_key,key_id,wrapped_dek,wrap_nonce,
		header_sha256,ciphertext_sha256,plaintext_sha256,ciphertext_bytes,
		plaintext_bytes,row_count,outcome,started_at
	) VALUES (p_export_id,p_attempt_generation,p_object_key,p_key_id,
		p_wrapped_dek,p_wrap_nonce,p_header_sha256,p_ciphertext_sha256,
		p_plaintext_sha256,p_ciphertext_bytes,p_plaintext_bytes,p_row_count,
		'started',p_started_at);
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_effect_start_v1(
    name,uuid,uuid,bigint,text,text,bytea,bytea,bytea,bytea,bytea,
    bigint,bigint,bigint,text,text,timestamptz
) FROM PUBLIC;
