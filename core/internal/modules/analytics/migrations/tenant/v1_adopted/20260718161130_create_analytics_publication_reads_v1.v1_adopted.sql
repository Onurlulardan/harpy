-- +goose Up
ALTER TABLE analytics_publication_releases
    ADD COLUMN requester_step_up_at timestamptz,
    ADD CONSTRAINT analytics_publication_requester_step_up_check CHECK (
        requester_step_up_at IS NULL OR pg_catalog.isfinite(requester_step_up_at)
    );

-- +goose StatementBegin
CREATE FUNCTION analytics_publication_get_v1(p_scope_schema name, p_release_id uuid)
RETURNS TABLE (
    release_id uuid, resource_kind text, resource_version_id uuid,
    audience text, lifecycle_state text, cache_policy text,
    manifest_sha256 bytea, requester_id uuid, approver_id uuid,
    row_revision bigint, generation bigint, public_id bytea,
    capability_expires_at timestamptz, created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY SELECT release.id, release.resource_kind, release.resource_version_id,
        release.audience, release.lifecycle_state, release.cache_policy,
        release.manifest_sha256, release.requester_id, release.approver_id,
        release.row_revision, release.generation, release.public_id,
        release.capability_expires_at, release.created_at, release.updated_at
    FROM analytics_publication_releases AS release WHERE release.id = p_release_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics publication was not found';
    END IF;
END;
$function$;

CREATE FUNCTION analytics_publication_list_v1(p_scope_schema name, p_limit integer, p_after uuid)
RETURNS TABLE (
    release_id uuid, resource_kind text, resource_version_id uuid,
    audience text, lifecycle_state text, cache_policy text,
    manifest_sha256 bytea, requester_id uuid, approver_id uuid,
    row_revision bigint, generation bigint, public_id bytea,
    capability_expires_at timestamptz, created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_limit IS NULL OR p_limit NOT BETWEEN 1 AND 100 THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics publication list input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY SELECT release.id, release.resource_kind, release.resource_version_id,
        release.audience, release.lifecycle_state, release.cache_policy,
        release.manifest_sha256, release.requester_id, release.approver_id,
        release.row_revision, release.generation, release.public_id,
        release.capability_expires_at, release.created_at, release.updated_at
    FROM analytics_publication_releases AS release
    WHERE p_after IS NULL OR release.id > p_after ORDER BY release.id LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_publication_get_v1(name,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_publication_list_v1(name,integer,uuid) FROM PUBLIC;
