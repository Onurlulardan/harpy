-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_project_dependencies_v1(p_target jsonb)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_version_id uuid;
    v_count bigint;
    v_matched bigint;
BEGIN
    v_version_id:=(p_target->'published_version'->>'id')::uuid;
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       pg_catalog.jsonb_typeof(p_target->'dependencies') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(p_target->'dependencies') NOT BETWEEN 1 AND 32
       OR EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_array_elements(p_target->'dependencies')
          WITH ORDINALITY AS dependency(value,position)
        WHERE (SELECT pg_catalog.array_agg(key ORDER BY key)
               FROM pg_catalog.jsonb_object_keys(dependency.value) AS key)
          IS DISTINCT FROM ARRAY[
            'catalog_version','created_at','data_revision_capability',
            'dependency_sha256','ordinal','owner_domain','owner_generation',
            'policy_version','query_contract_version','resource_id',
            'resource_kind','resource_version_id'
          ]::text[]
          OR (dependency.value->>'ordinal')::bigint<>dependency.position-1
          OR dependency.value->>'owner_domain' NOT IN ('iam','forms')
          OR dependency.value->>'dependency_sha256'!~'^[0-9a-f]{64}$'
       ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
          MESSAGE='Analytics V1 adoption dependency projection invalid';
    END IF;
    INSERT INTO analytics_query_dependencies(
      query_version_id,ordinal,owner_domain,resource_kind,resource_id,
      resource_version_id,owner_generation,catalog_version,
      query_contract_version,policy_version,data_revision_capability,
      dependency_sha256,created_at
    ) SELECT v_version_id,(dependency.value->>'ordinal')::smallint,
      dependency.value->>'owner_domain',dependency.value->>'resource_kind',
      (dependency.value->>'resource_id')::uuid,
      (dependency.value->>'resource_version_id')::uuid,
      (dependency.value->>'owner_generation')::bigint,
      (dependency.value->>'catalog_version')::bigint,
      (dependency.value->>'query_contract_version')::bigint,
      (dependency.value->>'policy_version')::bigint,
      dependency.value->>'data_revision_capability',
      pg_catalog.decode(dependency.value->>'dependency_sha256','hex'),
      (dependency.value->>'created_at')::timestamptz
    FROM pg_catalog.jsonb_array_elements(p_target->'dependencies')
      AS dependency(value)
    ON CONFLICT DO NOTHING;
    SELECT pg_catalog.count(*) INTO v_count
    FROM analytics_query_dependencies AS dependency
    WHERE dependency.query_version_id=v_version_id;
    SELECT pg_catalog.count(*) INTO v_matched
    FROM analytics_query_dependencies AS stored
    JOIN pg_catalog.jsonb_array_elements(p_target->'dependencies') AS expected(value)
      ON (expected.value->>'ordinal')::smallint=stored.ordinal
    WHERE stored.query_version_id=v_version_id
      AND stored.owner_domain=expected.value->>'owner_domain'
      AND stored.resource_kind=expected.value->>'resource_kind'
      AND stored.resource_id=(expected.value->>'resource_id')::uuid
      AND stored.resource_version_id=(expected.value->>'resource_version_id')::uuid
      AND stored.owner_generation=(expected.value->>'owner_generation')::bigint
      AND stored.catalog_version=(expected.value->>'catalog_version')::bigint
      AND stored.query_contract_version=(expected.value->>'query_contract_version')::bigint
      AND stored.policy_version=(expected.value->>'policy_version')::bigint
      AND stored.data_revision_capability=expected.value->>'data_revision_capability'
      AND stored.dependency_sha256=pg_catalog.decode(
        expected.value->>'dependency_sha256','hex');
    IF v_count IS DISTINCT FROM
       pg_catalog.jsonb_array_length(p_target->'dependencies')::bigint OR
       v_matched IS DISTINCT FROM v_count THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption dependency projection differs';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_project_dependencies_v1(jsonb)
FROM PUBLIC;
