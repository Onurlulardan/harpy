-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_dependency_sha256_v1(
    p_operation_id uuid,p_run_id uuid,p_tenant_id uuid
) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_stage analytics_v1_adoption_stage%ROWTYPE;
    v_manifest text;
    v_digest text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
      RAISE EXCEPTION USING ERRCODE='42501',
        MESSAGE='Analytics V1 adoption dependency evidence denied';
    END IF;
    FOR v_stage IN
      SELECT stage.* FROM analytics_v1_adoption_stage AS stage
      WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
        AND NOT stage.removed AND stage.disposition='adopted'
      ORDER BY stage.legacy_endpoint_id
    LOOP
      IF EXISTS (
        SELECT 1 FROM (
          SELECT dependency.value,
            pg_catalog.row_number() OVER (ORDER BY dependency.value->>'owner_domain',
              dependency.value->>'resource_kind',
              (dependency.value->>'resource_id')::uuid,
              (dependency.value->>'resource_version_id')::uuid)-1 AS expected
          FROM pg_catalog.jsonb_array_elements(
            v_stage.target_document->'dependencies'
          ) AS dependency(value)
        ) AS ordered
        WHERE (ordered.value->>'ordinal')::bigint<>ordered.expected OR
          ordered.value->>'dependency_sha256'=repeat('0',64) OR
          ((v_stage.target_document->'published_version'->>'cache_policy')=
            'revision_bound' AND
           pg_catalog.octet_length(
             ordered.value->>'data_revision_capability')=0) OR NOT EXISTS (
            SELECT 1 FROM platform.domain_cutovers AS owner
            WHERE owner.scope_kind='tenant' AND owner.scope_id=p_tenant_id
              AND owner.domain=ordered.value->>'owner_domain'
              AND owner.state='v2' AND owner.writer='v2'
              AND owner.generation=(ordered.value->>'owner_generation')::bigint
          )
      ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
          MESSAGE='Analytics V1 adoption dependency authority differs';
      END IF;
      SELECT pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
        pg_catalog.int4send(pg_catalog.octet_length(dependency.value->>'owner_domain'))||
        pg_catalog.convert_to(dependency.value->>'owner_domain','UTF8')||
        pg_catalog.int4send(pg_catalog.octet_length(dependency.value->>'resource_kind'))||
        pg_catalog.convert_to(dependency.value->>'resource_kind','UTF8')||
        pg_catalog.int4send(16)||pg_catalog.uuid_send(
          (dependency.value->>'resource_id')::uuid)||
        pg_catalog.int4send(16)||pg_catalog.uuid_send(
          (dependency.value->>'resource_version_id')::uuid)||
        pg_catalog.int4send(pg_catalog.octet_length(
          dependency.value->>'data_revision_capability'))||
        pg_catalog.convert_to(dependency.value->>'data_revision_capability','UTF8')||
        pg_catalog.int4send(32)||pg_catalog.decode(
          dependency.value->>'dependency_sha256','hex')||
        pg_catalog.int8send((dependency.value->>'catalog_version')::bigint)||
        pg_catalog.int8send((dependency.value->>'query_contract_version')::bigint)||
        pg_catalog.int8send((dependency.value->>'policy_version')::bigint),
        ''::bytea ORDER BY dependency.value->>'owner_domain',
          dependency.value->>'resource_kind',
          (dependency.value->>'resource_id')::uuid,
          (dependency.value->>'resource_version_id')::uuid),''::bytea)),'hex')
      INTO v_manifest FROM pg_catalog.jsonb_array_elements(
        v_stage.target_document->'dependencies'
      ) AS dependency(value);
      IF v_manifest IS DISTINCT FROM
         v_stage.target_document->'published_version'->>'manifest_sha256' OR
         v_manifest IS DISTINCT FROM
         v_stage.target_document->'draft_version'->>'manifest_sha256' THEN
        RAISE EXCEPTION USING ERRCODE='55000',
          MESSAGE='Analytics V1 adoption dependency manifest differs';
      END IF;
    END LOOP;
    SELECT pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
      pg_catalog.uuid_send(stage.legacy_endpoint_id)||stage.source_sha256||
      pg_catalog.decode(
        stage.target_document->'published_version'->>'manifest_sha256','hex'),
      ''::bytea ORDER BY stage.legacy_endpoint_id),''::bytea)),'hex')
    INTO v_digest FROM analytics_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND NOT stage.removed AND stage.disposition='adopted';
    RETURN v_digest;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_dependency_sha256_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
