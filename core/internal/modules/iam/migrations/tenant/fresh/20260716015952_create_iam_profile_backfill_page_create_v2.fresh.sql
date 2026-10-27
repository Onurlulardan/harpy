-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_backfill_page_create_v2(
    p_tenant_id uuid, p_plan_id uuid, p_resource_kind text,
    p_source_schema_id uuid, p_source_count bigint, p_page_size integer,
    p_expected_source bytea, p_expected_target bytea,
    p_initial_rolling bytea, p_pages jsonb, p_created_at timestamptz
)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_page_number integer;
    v_job_id uuid;
    v_payload_reference text;
    v_after_resource_id uuid;
    v_through_resource_id uuid;
    v_expected_count integer;
    v_source_sha256 bytea;
    v_expected_target_sha256 bytea;
    v_prior_through uuid;
    v_page_count integer;
    v_actual_count integer;
    v_actual_sha bytea;
    v_total_count bigint;
    v_source_evidence bytea;
    v_rolling bytea;
BEGIN
    v_source_evidence := ''::bytea;
    v_rolling := p_initial_rolling;
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR p_tenant_id IS NULL OR
       p_plan_id IS NULL OR p_resource_kind NOT IN ('user','organization','role') OR
       p_source_schema_id IS NULL OR p_source_count<0 OR p_page_size NOT BETWEEN 1 AND 250 OR
       pg_catalog.octet_length(p_expected_source)<>32 OR
       pg_catalog.octet_length(p_expected_target)<>32 OR
       pg_catalog.octet_length(p_initial_rolling)<>32 OR
       pg_catalog.jsonb_typeof(p_pages)<>'array' OR
       pg_catalog.jsonb_array_length(p_pages) NOT BETWEEN 1 AND 65536 OR EXISTS (
           SELECT 1 FROM pg_catalog.jsonb_array_elements(p_pages)
                WITH ORDINALITY AS page(value,ordinal)
           WHERE pg_catalog.jsonb_typeof(page.value)<>'object'
              OR (page.value->>'page_number')::integer<>page.ordinal
              OR (page.value->>'expected_count')::integer NOT BETWEEN 0 AND p_page_size
              OR pg_catalog.octet_length(pg_catalog.decode(
                    page.value->>'payload_sha256','hex'))<>32
              OR pg_catalog.octet_length(pg_catalog.decode(
                    page.value->>'source_sha256','hex'))<>32
              OR pg_catalog.octet_length(pg_catalog.decode(
                    page.value->>'expected_target_sha256','hex'))<>32
              OR EXISTS (
                   SELECT 1 FROM pg_catalog.jsonb_object_keys(page.value) AS key
                   WHERE key NOT IN ('page_number','job_id','scheduled_run_at',
                       'payload_reference','payload_sha256','after_resource_id',
                       'through_resource_id','expected_count','source_sha256',
                       'expected_target_sha256')
              )
       ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM profile backfill page input is invalid';
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
    GET DIAGNOSTICS v_page_count=ROW_COUNT;
    FOR v_page_number,v_job_id,v_payload_reference,v_after_resource_id,
        v_through_resource_id,v_expected_count,v_source_sha256,
        v_expected_target_sha256 IN
        SELECT page.page_number,page.job_id,page.payload_reference,
               page.after_resource_id,page.through_resource_id,page.expected_count,
               page.source_sha256,page.expected_target_sha256
        FROM iam_profile_backfill_pages AS page
        WHERE page.plan_id=p_plan_id ORDER BY page.page_number
    LOOP
        IF v_payload_reference<>pg_catalog.format(
               'urn:forgecrud:iam:profile-backfill:%s:%s:%s:v1',
               p_tenant_id,p_plan_id,v_page_number
           ) OR (v_page_number=1 AND v_after_resource_id IS NOT NULL) OR
           (v_page_number>1 AND v_after_resource_id<>v_prior_through) OR
           pg_catalog.substr(v_job_id::text,15,1)<>'7' THEN
            RAISE EXCEPTION USING ERRCODE='23514',
                MESSAGE='IAM profile backfill page identity is invalid';
        END IF;
        SELECT evidence.source_count,evidence.source_sha256
          INTO v_actual_count,v_actual_sha
        FROM iam_profile_backfill_source_page_v2(
            p_resource_kind,p_source_schema_id,v_after_resource_id,
            v_through_resource_id
        ) AS evidence;
        IF v_actual_count<>v_expected_count OR v_actual_sha<>v_source_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='40001',
                MESSAGE='IAM profile backfill source evidence changed';
        END IF;
        v_source_evidence:=v_source_evidence||v_source_sha256;
        v_rolling:=pg_catalog.sha256(v_rolling||v_expected_target_sha256);
        v_prior_through:=v_through_resource_id;
    END LOOP;
    SELECT pg_catalog.count(*) INTO v_total_count FROM (
        SELECT usr.id,usr.profile_schema_version_id FROM iam_users AS usr
        WHERE p_resource_kind='user'
        UNION ALL
        SELECT org.id,org.profile_schema_version_id FROM iam_organizations AS org
        WHERE p_resource_kind='organization'
        UNION ALL
        SELECT role.id,role.profile_schema_version_id FROM iam_roles AS role
        WHERE p_resource_kind='role'
    ) AS resource
    JOIN iam_profile_schema_versions AS row_schema
      ON row_schema.id=resource.profile_schema_version_id
     AND row_schema.resource_kind=p_resource_kind
     AND row_schema.state IN ('active','retired')
    JOIN iam_profile_schema_versions AS source_schema
      ON source_schema.id=p_source_schema_id
     AND source_schema.resource_kind=p_resource_kind
     AND source_schema.state='active'
     AND row_schema.version<=source_schema.version;
    IF v_total_count<>p_source_count OR
       (SELECT pg_catalog.sum(page.expected_count) FROM iam_profile_backfill_pages AS page
        WHERE page.plan_id=p_plan_id)<>p_source_count OR
       pg_catalog.sha256(v_source_evidence)<>p_expected_source OR
       v_rolling<>p_expected_target THEN
        RAISE EXCEPTION USING ERRCODE='23514',
            MESSAGE='IAM profile backfill aggregate evidence is invalid';
    END IF;
    RETURN v_page_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_backfill_page_create_v2(
    uuid,uuid,text,uuid,bigint,integer,bytea,bytea,bytea,jsonb,timestamptz
) FROM PUBLIC;
