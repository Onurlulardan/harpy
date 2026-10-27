-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_project_query_v1(p_target jsonb)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_query jsonb;
    v_published jsonb;
    v_draft jsonb;
BEGIN
    v_query:=p_target->'query';
    v_published:=p_target->'published_version';
    v_draft:=p_target->'draft_version';
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_query) AS key)
       IS DISTINCT FROM ARRAY[
        'created_at','created_by','current_draft_version_id','id',
        'lifecycle_state','reference_epoch','row_revision',
        'updated_at','updated_by'
       ]::text[] OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_published) AS key)
       IS DISTINCT FROM ARRAY[
        'cache_policy','created_at','created_by','id','manifest_sha256','name',
        'output_document','parameter_document','published_at','published_by',
        'query_document','query_id','row_revision','updated_at','updated_by',
        'version_no','version_state'
       ]::text[] OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_draft) AS key)
       IS DISTINCT FROM ARRAY[
        'cache_policy','created_at','created_by','id','manifest_sha256','name',
        'output_document','parameter_document','published_at','published_by',
        'query_document','query_id','row_revision','updated_at','updated_by',
        'version_no','version_state'
       ]::text[] OR
       v_query->>'lifecycle_state'<>'active' OR
       v_published->>'version_state'<>'published' OR
       v_draft->>'version_state'<>'draft' OR
       (v_query->>'id')::uuid IS DISTINCT FROM (v_published->>'query_id')::uuid OR
       (v_query->>'id')::uuid IS DISTINCT FROM (v_draft->>'query_id')::uuid OR
       (v_query->>'current_draft_version_id')::uuid
         IS DISTINCT FROM (v_draft->>'id')::uuid OR
       (v_published->>'id')::uuid=(v_draft->>'id')::uuid OR
       v_published->'published_by'='null'::jsonb OR
       v_published->'published_at'='null'::jsonb OR
       v_draft->'published_by'<>'null'::jsonb OR
       v_draft->'published_at'<>'null'::jsonb THEN
        RAISE EXCEPTION USING ERRCODE='22023',
          MESSAGE='Analytics V1 adoption query projection invalid';
    END IF;
    SET CONSTRAINTS analytics_saved_queries_current_draft_fk DEFERRED;
    INSERT INTO analytics_saved_queries(
      id,lifecycle_state,row_revision,reference_epoch,current_draft_version_id,
      created_by,updated_by,created_at,updated_at
    ) VALUES (
      (v_query->>'id')::uuid,v_query->>'lifecycle_state',
      (v_query->>'row_revision')::bigint,(v_query->>'reference_epoch')::bigint,
      (v_query->>'current_draft_version_id')::uuid,
      (v_query->>'created_by')::uuid,(v_query->>'updated_by')::uuid,
      (v_query->>'created_at')::timestamptz,(v_query->>'updated_at')::timestamptz
    ) ON CONFLICT DO NOTHING;
    INSERT INTO analytics_saved_query_versions(
      id,query_id,version_no,version_state,row_revision,name,
      query_document,query_sha256,parameter_document,parameter_sha256,
      output_document,output_sha256,manifest_sha256,cache_policy,
      created_by,updated_by,created_at,updated_at,published_by,published_at
    ) SELECT (row->>'id')::uuid,(row->>'query_id')::uuid,
      (row->>'version_no')::bigint,row->>'version_state',
      (row->>'row_revision')::bigint,row->>'name',
      pg_catalog.convert_to((row->'query_document')::text,'UTF8'),
      pg_catalog.sha256(pg_catalog.convert_to((row->'query_document')::text,'UTF8')),
      pg_catalog.convert_to((row->'parameter_document')::text,'UTF8'),
      pg_catalog.sha256(pg_catalog.convert_to((row->'parameter_document')::text,'UTF8')),
      pg_catalog.convert_to((row->'output_document')::text,'UTF8'),
      pg_catalog.sha256(pg_catalog.convert_to((row->'output_document')::text,'UTF8')),
      pg_catalog.decode(row->>'manifest_sha256','hex'),row->>'cache_policy',
      (row->>'created_by')::uuid,(row->>'updated_by')::uuid,
      (row->>'created_at')::timestamptz,(row->>'updated_at')::timestamptz,
      (row->>'published_by')::uuid,(row->>'published_at')::timestamptz
    FROM (VALUES(v_published),(v_draft)) AS value(row)
    ON CONFLICT DO NOTHING;
    PERFORM 1 FROM analytics_saved_queries AS root
    JOIN analytics_saved_query_versions AS published
      ON published.id=(v_published->>'id')::uuid AND published.query_id=root.id
    JOIN analytics_saved_query_versions AS draft
      ON draft.id=root.current_draft_version_id AND draft.query_id=root.id
    WHERE root.id=(v_query->>'id')::uuid
      AND root.row_revision=(v_query->>'row_revision')::bigint
      AND published.manifest_sha256=pg_catalog.decode(
        v_published->>'manifest_sha256','hex')
      AND published.query_document=pg_catalog.convert_to(
        (v_published->'query_document')::text,'UTF8')
      AND published.parameter_document=pg_catalog.convert_to(
        (v_published->'parameter_document')::text,'UTF8')
      AND published.output_document=pg_catalog.convert_to(
        (v_published->'output_document')::text,'UTF8')
      AND draft.version_state='draft';
    IF NOT FOUND THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption query projection differs';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_project_query_v1(jsonb) FROM PUBLIC;
