-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_public_approval_sha256_v1(
    p_operation_id uuid,p_run_id uuid,p_tenant_id uuid
) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_stage analytics_v1_adoption_stage%ROWTYPE;
    v_release jsonb;
    v_approval jsonb;
    v_snapshot jsonb;
    v_evidence bytea;
    v_digest text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
      RAISE EXCEPTION USING ERRCODE='42501',
        MESSAGE='Analytics V1 adoption public approval evidence denied';
    END IF;
    FOR v_stage IN
      SELECT stage.* FROM analytics_v1_adoption_stage AS stage
      WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
        AND NOT stage.removed AND stage.disposition='adopted'
      ORDER BY stage.legacy_endpoint_id
    LOOP
      IF (v_stage.source_document->>'is_public')::boolean IS NOT TRUE THEN
        IF v_stage.target_document->'release'<>'null'::jsonb OR
           v_stage.target_document->'approval'<>'null'::jsonb OR
           v_stage.target_document->'public_pointer'<>'null'::jsonb OR
           v_stage.target_document->'legacy_route'<>'null'::jsonb THEN
          RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics V1 adoption private publication differs';
        END IF;
        CONTINUE;
      END IF;
      v_release:=v_stage.target_document->'release';
      v_approval:=v_stage.target_document->'approval';
      v_evidence:=pg_catalog.sha256(
        pg_catalog.decode(v_approval->>'request_sha256','hex')||
        pg_catalog.decode(v_release->>'manifest_sha256','hex'));
      IF (v_approval->>'requester_id')::uuid=
           (v_approval->>'approver_id')::uuid OR
         v_approval->>'request_sha256'=repeat('0',64) OR
         (v_approval->>'approved_at')::timestamptz-
           (v_approval->>'requester_step_up_at')::timestamptz>
             interval '10 minutes' OR
         (v_approval->>'approved_at')::timestamptz-
           (v_approval->>'approver_step_up_at')::timestamptz>
             interval '10 minutes' OR
         pg_catalog.encode(v_evidence,'hex') IS DISTINCT FROM
           v_approval->>'evidence_sha256' THEN
        RAISE EXCEPTION USING ERRCODE='55000',
          MESSAGE='Analytics V1 adoption public approval differs';
      END IF;
      v_snapshot:=platform.analytics_v1_adoption_public_bridge_v1(
        'inspect',p_tenant_id,v_stage.legacy_endpoint_id,NULL);
      PERFORM 1 FROM analytics_publication_releases AS release
      JOIN analytics_publication_approvals AS approval
        ON approval.release_id=release.id
      WHERE release.id=(v_release->>'id')::uuid
        AND release.resource_version_id=
          (v_stage.target_document->'published_version'->>'id')::uuid
        AND release.audience='public' AND release.lifecycle_state='active'
        AND release.requester_id=(v_approval->>'requester_id')::uuid
        AND release.approver_id=(v_approval->>'approver_id')::uuid
        AND approval.evidence_sha256=v_evidence
        AND (v_snapshot->>'publication_exists')::boolean
        AND v_snapshot->>'publication_public_id'=
          pg_catalog.encode(release.public_id,'hex')
        AND (v_snapshot->>'publication_release_id')::uuid=release.id
        AND (v_snapshot->>'publication_resource_version_id')::uuid=
          release.resource_version_id
        AND v_snapshot->>'publication_audience'='public'
        AND v_snapshot->>'publication_state'='active';
      IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
          MESSAGE='Analytics V1 adoption stored public approval differs';
      END IF;
    END LOOP;
    SELECT pg_catalog.encode(pg_catalog.sha256(COALESCE(pg_catalog.string_agg(
      pg_catalog.uuid_send(stage.legacy_endpoint_id)||
      pg_catalog.uuid_send((stage.target_document->'release'->>'id')::uuid)||
      pg_catalog.decode(
        stage.target_document->'approval'->>'evidence_sha256','hex')||
      pg_catalog.decode(
        stage.target_document->'public_pointer'->>'public_id','hex'),
      ''::bytea ORDER BY stage.legacy_endpoint_id),''::bytea)),'hex')
    INTO v_digest FROM analytics_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND NOT stage.removed AND stage.disposition='adopted'
      AND (stage.source_document->>'is_public')::boolean;
    RETURN v_digest;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_public_approval_sha256_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
