-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_workflow_permission_scope_assert_v1(
    p_scope_schema name,
    p_tenant_id uuid
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_iam_group bigint;
    v_iam bigint;
    v_layout bigint;
    v_workflow_group bigint;
    v_workflow bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR NOT EXISTS (
           SELECT 1 FROM tenancy_schema_identity AS identity_row
           WHERE identity_row.singleton AND identity_row.tenant_id = p_tenant_id
             AND identity_row.schema_name = p_scope_schema::text
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Workflow permission IAM scope is invalid';
    END IF;
    SELECT group_row.generation, iam_row.generation, layout_row.generation
    INTO v_iam_group, v_iam, v_layout
    FROM platform.cutover_groups AS group_row
    JOIN platform.cutover_group_members AS iam_member
      ON iam_member.scope_kind=group_row.scope_kind
     AND iam_member.scope_id=group_row.scope_id
     AND iam_member.group_id=group_row.group_id
     AND iam_member.ordinal=1 AND iam_member.domain='iam'
    JOIN platform.cutover_group_members AS layout_member
      ON layout_member.scope_kind=group_row.scope_kind
     AND layout_member.scope_id=group_row.scope_id
     AND layout_member.group_id=group_row.group_id
     AND layout_member.ordinal=2 AND layout_member.domain='system_resource_layout'
    JOIN platform.domain_cutovers AS iam_row
      ON iam_row.scope_kind=group_row.scope_kind AND iam_row.scope_id=group_row.scope_id
     AND iam_row.domain='iam' AND iam_row.state='v2' AND iam_row.writer='v2'
    JOIN platform.domain_cutovers AS layout_row
      ON layout_row.scope_kind=group_row.scope_kind
     AND layout_row.scope_id=group_row.scope_id
     AND layout_row.domain='system_resource_layout'
     AND layout_row.state='v2' AND layout_row.writer='v2'
    WHERE group_row.scope_kind='tenant' AND group_row.scope_id=p_tenant_id
      AND group_row.group_id='iam_system_resources'
      AND group_row.state='v2' AND group_row.writer='v2'
      AND (SELECT pg_catalog.count(*) FROM platform.cutover_group_members AS extra
           WHERE extra.scope_kind=group_row.scope_kind
             AND extra.scope_id=group_row.scope_id
             AND extra.group_id=group_row.group_id)=2
    FOR SHARE OF group_row, iam_member, layout_member, iam_row, layout_row;
    SELECT group_row.generation, workflow_row.generation
    INTO v_workflow_group, v_workflow
    FROM platform.cutover_groups AS group_row
    JOIN platform.cutover_group_members AS member
      ON member.scope_kind=group_row.scope_kind AND member.scope_id=group_row.scope_id
     AND member.group_id=group_row.group_id
     AND member.ordinal=1 AND member.domain='workflow'
    JOIN platform.domain_cutovers AS workflow_row
      ON workflow_row.scope_kind=group_row.scope_kind
     AND workflow_row.scope_id=group_row.scope_id AND workflow_row.domain='workflow'
     AND workflow_row.state='v2' AND workflow_row.writer='v2'
    WHERE group_row.scope_kind='tenant' AND group_row.scope_id=p_tenant_id
      AND group_row.group_id='workflow'
      AND group_row.state='v2' AND group_row.writer='v2'
      AND (SELECT pg_catalog.count(*) FROM platform.cutover_group_members AS extra
           WHERE extra.scope_kind=group_row.scope_kind
             AND extra.scope_id=group_row.scope_id
             AND extra.group_id=group_row.group_id)=1
    FOR SHARE OF group_row, member, workflow_row;
    IF v_iam_group IS NULL OR v_iam_group IS DISTINCT FROM v_iam OR
       v_iam_group IS DISTINCT FROM v_layout OR v_workflow_group IS NULL OR
       v_workflow_group IS DISTINCT FROM v_workflow THEN
        RAISE EXCEPTION USING ERRCODE = 'P5731',
            MESSAGE = 'Workflow permission IAM authority is not current';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION iam_workflow_permission_scope_assert_v1(name,uuid)
    FROM PUBLIC;

CREATE FUNCTION iam_workflow_permission_authority_assert_v1(
    p_principal_id uuid,
    p_principal_revision bigint,
    p_principal_authorization_version bigint,
    p_grant_id uuid,
    p_grant_revision bigint,
    p_definition_id uuid,
    p_manifest_sha256 bytea,
    p_verified_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_principal_id IS NULL OR p_principal_revision < 1 OR
       p_principal_authorization_version < 1 OR p_grant_id IS NULL OR
       p_grant_revision < 1 OR p_definition_id IS NULL OR
       pg_catalog.octet_length(p_manifest_sha256) IS DISTINCT FROM 32 OR
       p_verified_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Workflow permission authority input is invalid';
    END IF;
    PERFORM 1 FROM iam_service_principals AS principal
    WHERE principal.id=p_principal_id AND principal.kind='workflow'
      AND principal.state='active' AND principal.revision=p_principal_revision
      AND principal.authorization_version=p_principal_authorization_version
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Workflow service principal is not current';
    END IF;
    PERFORM 1
    FROM iam_capability_grants AS head
    JOIN iam_capability_grant_revisions AS revision
      ON revision.grant_id=head.id
     AND revision.grant_revision=head.current_revision
     AND revision.principal_id=head.principal_id
     AND revision.capability_key=head.capability_key
    WHERE head.id=p_grant_id AND head.principal_id=p_principal_id
      AND head.capability_key='workflow.instance.execute'
      AND head.current_revision=p_grant_revision AND head.state='active'
      AND revision.disposition='active' AND revision.scope_kind='resource'
      AND revision.owner_domain='workflow' AND revision.resource_kind='definition'
      AND revision.resource_id=p_definition_id
      AND revision.manifest_sha256=p_manifest_sha256
      AND revision.valid_from<=p_verified_at
      AND (revision.valid_until IS NULL OR revision.valid_until>p_verified_at)
    FOR SHARE OF head, revision;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Workflow capability grant is not current';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION iam_workflow_permission_authority_assert_v1(
    uuid,bigint,bigint,uuid,bigint,uuid,bytea,timestamptz
) FROM PUBLIC;
-- +goose StatementEnd
