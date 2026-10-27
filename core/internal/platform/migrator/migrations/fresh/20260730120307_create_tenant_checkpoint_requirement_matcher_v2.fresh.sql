-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_tenant_checkpoint_satisfies_v2(
    p_tenant_id uuid,
    p_owner text,
    p_logical_version bigint,
    p_database_origin text,
    p_selected_path text,
    p_selected_sha256 text
)
RETURNS boolean
LANGUAGE sql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
SELECT EXISTS (
    SELECT 1
    FROM platform.tenant_migration_checkpoints AS checkpoint
    WHERE checkpoint.tenant_id = p_tenant_id
      AND checkpoint.owner = p_owner
      AND checkpoint.logical_version = p_logical_version
      AND checkpoint.state = 'succeeded'
      AND (
        (
          checkpoint.database_origin = p_database_origin
          AND checkpoint.selected_path = p_selected_path
          AND checkpoint.selected_sha256 = p_selected_sha256
        ) OR EXISTS (
          SELECT 1
          FROM platform.tenant_migration_supersessions AS supersession
          WHERE supersession.owner = p_owner
            AND supersession.scope = 'tenant'
            AND supersession.logical_version = p_logical_version
            AND supersession.database_origin = p_database_origin
            AND supersession.current_path = p_selected_path
            AND supersession.current_sha256 = p_selected_sha256
            AND checkpoint.database_origin = supersession.database_origin
            AND checkpoint.selected_path = supersession.superseded_path
            AND checkpoint.selected_sha256 = supersession.superseded_sha256
            AND supersession.supersession_distance IN (1,2)
        )
      )
);
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_migrator_tenant_checkpoint_satisfies_v2(
        uuid,text,bigint,text,text,text
    )
FROM PUBLIC;
