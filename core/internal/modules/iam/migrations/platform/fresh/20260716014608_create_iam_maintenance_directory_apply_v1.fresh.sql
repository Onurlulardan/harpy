-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_maintenance_directory_apply_v1(
    p_tenant_id uuid, p_identifier_id uuid, p_user_id uuid,
    p_kind text, p_state text, p_source_revision bigint,
    p_source_sha256 bytea, p_key_evidence jsonb
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_entry jsonb;
    v_key_id text;
    v_hmac bytea;
    v_rows bigint;
    v_step bigint;
BEGIN
    v_rows := 0;
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_tenant_id IS NULL OR p_identifier_id IS NULL OR p_user_id IS NULL OR
       p_kind NOT IN ('email_v1', 'portal_v1') OR
       p_state NOT IN ('pending', 'verified', 'active', 'tombstoned', 'retired') OR
       p_source_revision < 1 OR
       pg_catalog.octet_length(p_source_sha256) IS DISTINCT FROM 32 OR
       NOT EXISTS (
           SELECT 1 FROM platform.tenants AS tenant
           WHERE tenant.id = p_tenant_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM maintenance directory input is invalid';
    END IF;
    IF p_key_evidence IS NULL OR
       pg_catalog.jsonb_typeof(p_key_evidence) IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM maintenance directory evidence is invalid';
    END IF;
    IF pg_catalog.jsonb_array_length(p_key_evidence) > 65 OR
       (p_state = 'active' AND
        pg_catalog.jsonb_array_length(p_key_evidence) < 1) OR
       (p_state <> 'active' AND
        pg_catalog.jsonb_array_length(p_key_evidence) <> 0) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM maintenance directory evidence is invalid';
    END IF;
    IF p_state = 'active' THEN
        IF EXISTS (
            SELECT 1
            FROM pg_catalog.jsonb_array_elements(p_key_evidence) AS item(value)
            WHERE pg_catalog.jsonb_typeof(item.value) <> 'object'
               OR item.value <> pg_catalog.jsonb_build_object(
                    'key_id', item.value ->> 'key_id',
                    'hmac', item.value ->> 'hmac'
                  )
               OR NOT platform.iam_hmac_key_id_valid_v1(
                    item.value ->> 'key_id'
                  )
               OR item.value ->> 'hmac' !~ '^[0-9a-f]{64}$'
        ) OR (
            SELECT pg_catalog.count(DISTINCT item.value ->> 'key_id')
            FROM pg_catalog.jsonb_array_elements(p_key_evidence) AS item(value)
        ) <> pg_catalog.jsonb_array_length(p_key_evidence) THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'IAM maintenance directory evidence is invalid';
        END IF;
        FOR v_entry IN
            SELECT item.value
            FROM pg_catalog.jsonb_array_elements(p_key_evidence) AS item(value)
            ORDER BY item.value ->> 'key_id'
        LOOP
            v_key_id := v_entry ->> 'key_id';
            v_hmac := pg_catalog.decode(v_entry ->> 'hmac', 'hex');
            UPDATE platform.iam_login_directory AS directory
            SET identifier_kind = p_kind, identifier_hmac = v_hmac,
                user_id = p_user_id, source_revision = p_source_revision,
                state = 'active', source_sha256 = p_source_sha256,
                projected_at = v_now, tombstoned_at = NULL
            WHERE directory.tenant_id = p_tenant_id
              AND directory.lookup_key_id = v_key_id
              AND directory.source_identifier_id = p_identifier_id;
            GET DIAGNOSTICS v_step = ROW_COUNT;
            IF v_step = 0 THEN
                INSERT INTO platform.iam_login_directory (
                    tenant_id, identifier_kind, lookup_key_id, identifier_hmac,
                    user_id, source_identifier_id, source_revision, state,
                    source_sha256, projected_at, tombstoned_at
                ) VALUES (
                    p_tenant_id, p_kind, v_key_id, v_hmac, p_user_id,
                    p_identifier_id, p_source_revision, 'active',
                    p_source_sha256, v_now, NULL
                );
                v_step := 1;
            END IF;
            v_rows := v_rows + v_step;
        END LOOP;
        UPDATE platform.iam_login_directory AS directory
        SET state = 'tombstoned', source_revision = p_source_revision,
            source_sha256 = p_source_sha256, projected_at = v_now,
            tombstoned_at = COALESCE(directory.tombstoned_at, v_now)
        WHERE directory.tenant_id = p_tenant_id
          AND directory.source_identifier_id = p_identifier_id
          AND NOT EXISTS (
              SELECT 1
              FROM pg_catalog.jsonb_array_elements(p_key_evidence) AS item(value)
              WHERE item.value ->> 'key_id' = directory.lookup_key_id
          );
        GET DIAGNOSTICS v_step = ROW_COUNT;
        v_rows := v_rows + v_step;
    ELSE
        UPDATE platform.iam_login_directory AS directory
        SET state = 'tombstoned', source_revision = p_source_revision,
            source_sha256 = p_source_sha256, projected_at = v_now,
            tombstoned_at = COALESCE(directory.tombstoned_at, v_now)
        WHERE directory.tenant_id = p_tenant_id
          AND directory.source_identifier_id = p_identifier_id
          AND directory.source_revision <= p_source_revision;
        GET DIAGNOSTICS v_rows = ROW_COUNT;
    END IF;
    RETURN v_rows;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_maintenance_directory_apply_v1(
    uuid,uuid,uuid,text,text,bigint,bytea,jsonb
) FROM PUBLIC;
