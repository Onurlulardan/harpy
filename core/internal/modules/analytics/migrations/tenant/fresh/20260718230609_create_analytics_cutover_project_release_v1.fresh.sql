-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_project_release_v1(p_target jsonb)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_release jsonb;
    v_approval jsonb;
BEGIN
    v_release:=p_target->'release';
    v_approval:=p_target->'approval';
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
      RAISE EXCEPTION USING ERRCODE='42501',
        MESSAGE='Analytics V1 adoption release projection principal denied';
    END IF;
    IF (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_release) AS key)
       IS DISTINCT FROM ARRAY[
        'activated_at','approved_at','approver_id','audience','cache_policy',
        'capability_expires_at','created_at','generation','id','lifecycle_state',
        'manifest_sha256','public_id','requester_id','requester_step_up_at',
        'resource_kind','resource_version_id','revoked_at','row_revision',
        'submitted_at','updated_at'
       ]::text[] OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_approval) AS key)
       IS DISTINCT FROM ARRAY[
        'approved_at','approver_id','approver_step_up_at','evidence_sha256',
        'release_id','requester_id','requester_step_up_at','request_sha256'
       ]::text[] THEN
      RAISE EXCEPTION USING ERRCODE='22023',
        MESSAGE='Analytics V1 adoption release projection shape invalid';
    END IF;
    IF
       v_release->>'resource_kind'<>'query' OR
       v_release->>'audience'<>'public' OR
       v_release->>'lifecycle_state'<>'active' OR
       v_release->>'cache_policy' NOT IN ('none','revision_bound') OR
       v_release->>'manifest_sha256'!~'^[0-9a-f]{64}$' OR
       v_release->>'public_id'!~'^[0-9a-f]{32}$' OR
       (v_release->>'resource_version_id')::uuid IS DISTINCT FROM
         (p_target->'published_version'->>'id')::uuid OR
       v_release->>'manifest_sha256' IS DISTINCT FROM
         p_target->'published_version'->>'manifest_sha256' OR
       (v_approval->>'release_id')::uuid IS DISTINCT FROM
         (v_release->>'id')::uuid OR
       (v_approval->>'requester_id')::uuid IS DISTINCT FROM
         (v_release->>'requester_id')::uuid OR
       (v_approval->>'approver_id')::uuid IS DISTINCT FROM
         (v_release->>'approver_id')::uuid OR
       v_approval->>'request_sha256'!~'^[0-9a-f]{64}$' OR
       v_approval->>'evidence_sha256'!~'^[0-9a-f]{64}$' THEN
      RAISE EXCEPTION USING ERRCODE='22023',
        MESSAGE='Analytics V1 adoption release projection invalid';
    END IF;
    INSERT INTO analytics_publication_releases(
      id,resource_kind,resource_version_id,audience,lifecycle_state,cache_policy,
      manifest_sha256,requester_id,approver_id,row_revision,generation,public_id,
      capability_expires_at,created_at,updated_at,submitted_at,approved_at,
      requester_step_up_at,
      activated_at,revoked_at
    ) VALUES (
      (v_release->>'id')::uuid,v_release->>'resource_kind',
      (v_release->>'resource_version_id')::uuid,v_release->>'audience',
      v_release->>'lifecycle_state',v_release->>'cache_policy',
      pg_catalog.decode(v_release->>'manifest_sha256','hex'),
      (v_release->>'requester_id')::uuid,(v_release->>'approver_id')::uuid,
      (v_release->>'row_revision')::bigint,(v_release->>'generation')::bigint,
      pg_catalog.decode(v_release->>'public_id','hex'),NULL,
      (v_release->>'created_at')::timestamptz,(v_release->>'updated_at')::timestamptz,
      (v_release->>'submitted_at')::timestamptz,
      (v_release->>'approved_at')::timestamptz,
      (v_release->>'requester_step_up_at')::timestamptz,
      (v_release->>'activated_at')::timestamptz,NULL
    ) ON CONFLICT DO NOTHING;
    INSERT INTO analytics_publication_approvals(
      release_id,requester_id,requester_step_up_at,approver_id,
      approver_step_up_at,approved_at,evidence_sha256
    ) VALUES (
      (v_approval->>'release_id')::uuid,(v_approval->>'requester_id')::uuid,
      (v_approval->>'requester_step_up_at')::timestamptz,
      (v_approval->>'approver_id')::uuid,
      (v_approval->>'approver_step_up_at')::timestamptz,
      (v_approval->>'approved_at')::timestamptz,
      pg_catalog.decode(v_approval->>'evidence_sha256','hex')
    ) ON CONFLICT DO NOTHING;
    PERFORM 1 FROM analytics_publication_releases AS release
    JOIN analytics_publication_approvals AS approval ON approval.release_id=release.id
    WHERE release.id=(v_release->>'id')::uuid
      AND release.resource_version_id=(v_release->>'resource_version_id')::uuid
      AND release.audience='public' AND release.lifecycle_state='active'
      AND release.manifest_sha256=pg_catalog.decode(v_release->>'manifest_sha256','hex')
      AND release.requester_id=(v_release->>'requester_id')::uuid
      AND release.approver_id=(v_release->>'approver_id')::uuid
      AND release.requester_step_up_at=
        (v_approval->>'requester_step_up_at')::timestamptz
      AND release.public_id=pg_catalog.decode(v_release->>'public_id','hex')
      AND approval.requester_id=release.requester_id
      AND approval.approver_id=release.approver_id
      AND approval.evidence_sha256=pg_catalog.decode(
        v_approval->>'evidence_sha256','hex')
      AND approval.evidence_sha256=pg_catalog.sha256(
        pg_catalog.decode(v_approval->>'request_sha256','hex')||
        release.manifest_sha256);
    IF NOT FOUND THEN
      RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption release projection differs';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_project_release_v1(jsonb)
FROM PUBLIC;
