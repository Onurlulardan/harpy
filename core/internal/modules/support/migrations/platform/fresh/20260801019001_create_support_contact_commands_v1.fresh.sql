-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.support_submit_contact_v1(
    p_id uuid, p_name text, p_email text, p_company text, p_phone text,
    p_employee_count text, p_message text, p_contact_type text, p_source text,
    p_rate_key_sha256 bytea, p_idempotency_key text, p_request_sha256 bytea,
    p_observed_at timestamp(6) with time zone
) RETURNS TABLE(response_document jsonb, replayed boolean, admitted boolean)
LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_partition text;
    v_receipt platform.support_command_receipts%ROWTYPE;
    v_message platform.support_contact_messages%ROWTYPE;
    v_count integer;
    v_bucket timestamp(6) with time zone;
    v_document jsonb;
BEGIN
    IF p_id IS NULL OR pg_catalog.uuid_extract_version(p_id) IS DISTINCT FROM 7 OR
       p_name IS NULL OR p_name <> pg_catalog.btrim(p_name) OR pg_catalog.octet_length(p_name) NOT BETWEEN 1 AND 200 OR
       p_email IS NULL OR p_email <> pg_catalog.lower(pg_catalog.btrim(p_email)) OR pg_catalog.octet_length(p_email) NOT BETWEEN 3 AND 255 OR
       p_company IS NULL OR p_company <> pg_catalog.btrim(p_company) OR pg_catalog.octet_length(p_company) > 200 OR
       p_phone IS NULL OR p_phone <> pg_catalog.btrim(p_phone) OR pg_catalog.octet_length(p_phone) > 50 OR
       p_employee_count IS NULL OR p_employee_count NOT IN ('','1-10','11-50','51-200','201-500','500+') OR
       p_message IS NULL OR p_message <> pg_catalog.btrim(p_message) OR pg_catalog.octet_length(p_message) NOT BETWEEN 1 AND 2000 OR
       p_contact_type IS NULL OR p_contact_type NOT IN ('general','demo_request','pricing_inquiry','technical_question','partnership','enterprise_plan','on_premise') OR
       p_source IS NULL OR p_source <> pg_catalog.btrim(p_source) OR pg_catalog.octet_length(p_source) > 50 OR
       p_rate_key_sha256 IS NULL OR pg_catalog.octet_length(p_rate_key_sha256) <> 32 OR
       p_idempotency_key IS NULL OR pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 16 AND 200 OR
       p_request_sha256 IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) OR
       pg_catalog.date_trunc('microseconds', p_observed_at) <> p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6501', MESSAGE = 'support contact command invalid';
    END IF;

    v_partition := 'contact:' || pg_catalog.encode(p_rate_key_sha256, 'hex');
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'support.contact.create:' || v_partition || ':' || p_idempotency_key, 0));
    SELECT receipt.* INTO v_receipt
    FROM platform.support_command_receipts AS receipt
    WHERE receipt.operation = 'support.contact.create'
      AND receipt.principal_partition = v_partition
      AND receipt.idempotency_key = p_idempotency_key;
    IF FOUND THEN
        IF v_receipt.request_sha256 <> p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6502', MESSAGE = 'support idempotency mismatch';
        END IF;
        RETURN QUERY SELECT v_receipt.response_document, true, true;
        RETURN;
    END IF;

    v_bucket := pg_catalog.date_bin('15 minutes'::interval, p_observed_at,
        '2000-01-01 00:00:00+00'::timestamp with time zone);
    INSERT INTO platform.support_public_rate_leases AS lease(
        rate_key_sha256, bucket_start, request_count, updated_at)
    VALUES (p_rate_key_sha256, v_bucket, 1, p_observed_at)
    ON CONFLICT (rate_key_sha256, bucket_start) DO UPDATE
    SET request_count = lease.request_count + 1, updated_at = EXCLUDED.updated_at
    RETURNING request_count INTO v_count;
    IF v_count > 5 THEN
        RETURN QUERY SELECT NULL::jsonb, false, false;
        RETURN;
    END IF;

    INSERT INTO platform.support_contact_messages(
        id, name, email, company, phone, employee_count, message, contact_type,
        status, source, revision, created_at, updated_at)
    VALUES (p_id, p_name, p_email, p_company, p_phone, p_employee_count, p_message,
        p_contact_type, 'new', p_source, 1, p_observed_at, p_observed_at)
    RETURNING * INTO v_message;
    v_document := pg_catalog.to_jsonb(v_message);
    INSERT INTO platform.support_command_receipts(
        receipt_id, operation, principal_partition, idempotency_key,
        request_sha256, result_id, result_revision, response_document, created_at)
    VALUES (pg_catalog.uuidv7(), 'support.contact.create', v_partition,
        p_idempotency_key, p_request_sha256, p_id, 1, v_document, p_observed_at);
    RETURN QUERY SELECT v_document, false, true;
END;
$function$;

CREATE FUNCTION platform.support_update_contact_v1(
    p_actor_id uuid, p_authorization_version bigint, p_authority_clock bigint,
    p_id uuid, p_has_status boolean, p_status text, p_has_notes boolean, p_notes text,
    p_expected_revision bigint, p_idempotency_key text, p_request_sha256 bytea,
    p_observed_at timestamp(6) with time zone
) RETURNS TABLE(response_document jsonb, replayed boolean)
LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_partition text;
    v_receipt platform.support_command_receipts%ROWTYPE;
    v_message platform.support_contact_messages%ROWTYPE;
    v_document jsonb;
BEGIN
    IF p_actor_id IS NULL OR p_authorization_version < 1 OR p_authority_clock < 1 OR
       p_id IS NULL OR (NOT p_has_status AND NOT p_has_notes) OR
       (p_has_status AND (p_status IS NULL OR p_status NOT IN ('new','read','replied','archived'))) OR
       (NOT p_has_status AND p_status IS NOT NULL) OR
       (p_has_notes AND (p_notes IS NULL OR pg_catalog.octet_length(p_notes) > 8000)) OR
       (NOT p_has_notes AND p_notes IS NOT NULL) OR p_expected_revision < 1 OR
       p_idempotency_key IS NULL OR pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 16 AND 200 OR
       p_request_sha256 IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) OR
       pg_catalog.date_trunc('microseconds', p_observed_at) <> p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6501', MESSAGE = 'support contact update invalid';
    END IF;
    v_partition := 'actor:' || p_actor_id::text;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'support.contact.update:' || v_partition || ':' || p_idempotency_key, 0));
    SELECT receipt.* INTO v_receipt FROM platform.support_command_receipts AS receipt
    WHERE receipt.operation = 'support.contact.update'
      AND receipt.principal_partition = v_partition
      AND receipt.idempotency_key = p_idempotency_key;
    IF FOUND THEN
        IF v_receipt.request_sha256 <> p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6502', MESSAGE = 'support idempotency mismatch';
        END IF;
        RETURN QUERY SELECT v_receipt.response_document, true;
        RETURN;
    END IF;
    SELECT message.* INTO v_message FROM platform.support_contact_messages AS message
    WHERE message.id = p_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6504', MESSAGE = 'support contact missing';
    END IF;
    IF v_message.revision <> p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = 'P6503', MESSAGE = 'support revision stale';
    END IF;
    UPDATE platform.support_contact_messages AS message
    SET status = CASE WHEN p_has_status THEN p_status ELSE message.status END,
        notes = CASE WHEN p_has_notes THEN p_notes ELSE message.notes END,
        replied_at = CASE WHEN p_has_status AND p_status = 'replied' THEN p_observed_at ELSE message.replied_at END,
        replied_by = CASE WHEN p_has_status AND p_status = 'replied' THEN p_actor_id ELSE message.replied_by END,
        revision = message.revision + 1, updated_at = p_observed_at
    WHERE message.id = p_id RETURNING * INTO v_message;
    v_document := pg_catalog.to_jsonb(v_message);
    INSERT INTO platform.support_command_receipts(
        receipt_id, operation, principal_partition, idempotency_key,
        request_sha256, result_id, result_revision, response_document, created_at)
    VALUES (pg_catalog.uuidv7(), 'support.contact.update', v_partition,
        p_idempotency_key, p_request_sha256, p_id, v_message.revision,
        v_document, p_observed_at);
    RETURN QUERY SELECT v_document, false;
END;
$function$;

REVOKE ALL ON FUNCTION platform.support_submit_contact_v1(uuid,text,text,text,text,text,text,text,text,bytea,text,bytea,timestamp with time zone) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.support_update_contact_v1(uuid,bigint,bigint,uuid,boolean,text,boolean,text,bigint,text,bytea,timestamp with time zone) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_submit_contact_v1(uuid,text,text,text,text,text,text,text,text,bytea,text,bytea,timestamp with time zone) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.support_update_contact_v1(uuid,bigint,bigint,uuid,boolean,text,boolean,text,bigint,text,bytea,timestamp with time zone) TO forgecrud_runtime;
-- +goose StatementEnd
