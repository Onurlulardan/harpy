-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_recovery_pages_create_v1(
    p_tenant_id uuid, p_plan_id uuid, p_resource_kind text,
    p_source_schema_id uuid, p_target_schema_id uuid,
    p_remaining_count bigint, p_page_size integer,
    p_remaining_source_sha256 bytea, p_expected_target_sha256 bytea,
    p_initial_rolling_sha256 bytea, p_inherited_target_count bigint,
    p_inherited_target_sha256 bytea, p_pages jsonb, p_created_at timestamptz
)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_number integer;
    v_job_id uuid;
    v_reference text;
    v_after uuid;
    v_through uuid;
    v_expected_count integer;
    v_source_sha256 bytea;
    v_target_sha256 bytea;
    v_prior uuid;
    v_page_count integer;
    v_actual_count integer;
    v_actual_sha256 bytea;
    v_target_count bigint;
    v_inherited_sha256 bytea;
    v_source_evidence bytea;
    v_rolling bytea;
    v_total bigint;
BEGIN
    v_source_evidence := pg_catalog.sha256(''::bytea);
    v_rolling := p_initial_rolling_sha256;
    v_total := 0;
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR p_tenant_id IS NULL OR
       p_plan_id IS NULL OR p_resource_kind NOT IN ('user','organization','role') OR
       p_source_schema_id IS NULL OR p_target_schema_id IS NULL OR
       p_source_schema_id = p_target_schema_id OR p_remaining_count < 1 OR
       p_page_size NOT BETWEEN 1 AND 250 OR p_inherited_target_count < 0 OR
       pg_catalog.octet_length(p_remaining_source_sha256) <> 32 OR
       pg_catalog.octet_length(p_expected_target_sha256) <> 32 OR
       pg_catalog.octet_length(p_initial_rolling_sha256) <> 32 OR
       pg_catalog.octet_length(p_inherited_target_sha256) <> 32 OR
       pg_catalog.jsonb_typeof(p_pages) <> 'array' OR
       pg_catalog.jsonb_array_length(p_pages) NOT BETWEEN 1 AND 65536 OR EXISTS (
           SELECT 1 FROM pg_catalog.jsonb_array_elements(p_pages)
                WITH ORDINALITY AS page(value,ordinal)
           WHERE pg_catalog.jsonb_typeof(page.value) <> 'object'
              OR (page.value->>'page_number')::integer <> page.ordinal
              OR (page.value->>'expected_count')::integer NOT BETWEEN 1 AND p_page_size
              OR pg_catalog.octet_length(pg_catalog.decode(
                    page.value->>'payload_sha256','hex')) <> 32
              OR pg_catalog.octet_length(pg_catalog.decode(
                    page.value->>'source_sha256','hex')) <> 32
              OR pg_catalog.octet_length(pg_catalog.decode(
                    page.value->>'expected_target_sha256','hex')) <> 32
              OR EXISTS (SELECT 1 FROM pg_catalog.jsonb_object_keys(page.value) AS key
                 WHERE key NOT IN ('page_number','job_id','scheduled_run_at',
                     'payload_reference','payload_sha256','after_resource_id',
                     'through_resource_id','expected_count','source_sha256',
                     'expected_target_sha256'))
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile recovery page input is invalid';
    END IF;
    INSERT INTO iam_profile_backfill_pages (
        plan_id,page_number,job_id,scheduled_run_at,payload_reference,payload_sha256,
        after_resource_id,through_resource_id,expected_count,source_sha256,
        expected_target_sha256,state,converted_count,error_count,result_sha256,
        completed_at,created_at
    )
    SELECT p_plan_id,page.ordinal,(page.value->>'job_id')::uuid,
        (page.value->>'scheduled_run_at')::timestamptz,page.value->>'payload_reference',
        pg_catalog.decode(page.value->>'payload_sha256','hex'),
        (page.value->>'after_resource_id')::uuid,(page.value->>'through_resource_id')::uuid,
        (page.value->>'expected_count')::integer,
        pg_catalog.decode(page.value->>'source_sha256','hex'),
        pg_catalog.decode(page.value->>'expected_target_sha256','hex'),
        'queued',0,0,NULL,NULL,p_created_at
    FROM pg_catalog.jsonb_array_elements(p_pages)
         WITH ORDINALITY AS page(value,ordinal);
    GET DIAGNOSTICS v_page_count = ROW_COUNT;
    FOR v_number,v_job_id,v_reference,v_after,v_through,v_expected_count,
        v_source_sha256,v_target_sha256 IN
        SELECT page.page_number,page.job_id,page.payload_reference,
               page.after_resource_id,page.through_resource_id,
               page.expected_count,page.source_sha256,page.expected_target_sha256
        FROM iam_profile_backfill_pages AS page
        WHERE page.plan_id = p_plan_id ORDER BY page.page_number
    LOOP
        IF v_reference <> pg_catalog.format(
               'urn:harpy:iam:profile-backfill:%s:%s:%s:v1',
               p_tenant_id,p_plan_id,v_number
           ) OR (v_number = 1 AND v_after IS NOT NULL) OR
           (v_number > 1 AND v_after <> v_prior) OR
           pg_catalog.substr(v_job_id::text,15,1) <> '7' THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'IAM profile recovery page identity is invalid';
        END IF;
        SELECT evidence.source_count,evidence.source_sha256
        INTO v_actual_count,v_actual_sha256
        FROM iam_profile_recovery_source_page_v1(
            p_resource_kind,p_source_schema_id,p_target_schema_id,v_after,v_through
        ) AS evidence;
        IF v_actual_count <> v_expected_count OR
           v_actual_sha256 <> v_source_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM profile recovery source evidence changed';
        END IF;
        v_source_evidence := pg_catalog.sha256(
            v_source_evidence || v_source_sha256
        );
        v_rolling := pg_catalog.sha256(v_rolling || v_target_sha256);
        v_total := v_total + v_expected_count;
        v_prior := v_through;
    END LOOP;
    IF EXISTS (
        SELECT 1 FROM iam_profile_recovery_resources_v1 AS resource
        WHERE resource.resource_kind = p_resource_kind
          AND resource.schema_version_id <> p_target_schema_id
          AND resource.resource_id > v_prior
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile recovery source evidence changed';
    END IF;
    SELECT evidence.target_count,evidence.target_sha256
    INTO v_target_count,v_inherited_sha256
    FROM iam_profile_recovery_target_evidence_v1(
        p_resource_kind,p_target_schema_id
    ) AS evidence;
    IF v_total <> p_remaining_count OR
        v_source_evidence <> p_remaining_source_sha256 OR
        v_rolling <> p_expected_target_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile recovery aggregate evidence is invalid';
    END IF;
    IF v_target_count <> p_inherited_target_count OR
       v_inherited_sha256 <> p_inherited_target_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile recovery target evidence changed';
    END IF;
    RETURN v_page_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_recovery_pages_create_v1(
    uuid,uuid,text,uuid,uuid,bigint,integer,bytea,bytea,bytea,bigint,bytea,jsonb,timestamptz
) FROM PUBLIC;
