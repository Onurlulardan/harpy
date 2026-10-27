-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_capability_grant_read_v1(
    p_scope_schema name,
    p_grant_id uuid,
    p_grant_revision bigint
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_grant_id IS NULL OR p_grant_revision IS NULL OR p_grant_revision < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM capability grant read scope is invalid';
    END IF;
    SELECT pg_catalog.jsonb_build_object(
        'principal', pg_catalog.jsonb_build_object(
            'id', principal.id,
            'kind', principal.kind,
            'state', principal.state,
            'authorization_version', principal.authorization_version,
            'revision', principal.revision
        ),
        'grant', pg_catalog.jsonb_build_object(
            'id', head.id,
            'revision', revision.grant_revision,
            'previous_revision', COALESCE(revision.previous_revision, 0),
            'principal_id', revision.principal_id,
            'capability', revision.capability_key,
            'scope_kind', revision.scope_kind,
            'owner', revision.owner_domain,
            'resource_kind', revision.resource_kind,
            'resource_id', revision.resource_id,
            'valid_from', revision.valid_from,
            'valid_until', revision.valid_until,
            'state', revision.disposition,
            'manifest_sha256', pg_catalog.encode(revision.manifest_sha256, 'hex')
        )
    )
    INTO v_document
    FROM iam_capability_grants AS head
    JOIN iam_capability_grant_revisions AS revision
      ON revision.grant_id = head.id
     AND revision.grant_revision = head.current_revision
    JOIN iam_service_principals AS principal
      ON principal.id = head.principal_id
    WHERE head.id = p_grant_id
      AND head.current_revision = p_grant_revision
      AND revision.principal_id = head.principal_id
      AND revision.capability_key = head.capability_key;
    RETURN v_document;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_capability_grant_read_v1(name,uuid,bigint)
    FROM PUBLIC;
