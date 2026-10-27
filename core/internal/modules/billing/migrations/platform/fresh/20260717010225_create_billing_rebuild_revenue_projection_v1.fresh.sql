-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_rebuild_revenue_projection_v1(
    p_rebuild_id uuid,
    p_actor_id uuid,
    p_authority_id uuid,
    p_rebuilt_at timestamptz
)
RETURNS TABLE (
    source_count bigint,
    projected_count bigint,
    source_revision_ceiling bigint,
    source_sha256 bytea,
    projection_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    group_generation bigint;
    billing_generation bigint;
    before_sha bytea;
    before_count bigint;
    before_revision bigint;
    source_count_value bigint;
    source_revision_value bigint;
    source_sha_value bytea;
    projection_count_value bigint;
    projection_revision_value bigint;
    projection_sha_value bytea;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR pg_catalog.uuid_extract_version(p_rebuild_id) <> 7
       OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR p_rebuilt_at IS DISTINCT FROM
          pg_catalog.date_trunc('microseconds', p_rebuilt_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing revenue rebuild input is invalid';
    END IF;

    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    SELECT authority.authority_group_generation,
           authority.authority_domain_generation
    INTO group_generation, billing_generation
    FROM platform.billing_global_authorities AS authority
    WHERE authority.authority_id = p_authority_id
    FOR SHARE;

    IF group_generation IS NULL OR billing_generation IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing revenue rebuild authority is unavailable';
    END IF;

    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'forgecrud:v2:billing:revenue-rebuild', 0
    ));

    SELECT digest.row_count, digest.revision_ceiling,
           digest.canonical_sha256
    INTO before_count, before_revision, before_sha
    FROM platform.billing_revenue_projection_digest_v1(p_authority_id) AS digest;

    SELECT digest.row_count, digest.revision_ceiling,
           digest.canonical_sha256
    INTO source_count_value, source_revision_value, source_sha_value
    FROM platform.billing_revenue_source_digest_v1(p_authority_id) AS digest;

    DELETE FROM platform.billing_revenue_projection;

    PERFORM platform.billing_refresh_revenue_projection_v1(source.id)
    FROM (
        SELECT subscription.id
        FROM platform.billing_subscriptions AS subscription
        JOIN platform.billing_price_versions AS price
          ON price.id = subscription.price_version_id
         AND price.offering_version_id = subscription.offering_version_id
        WHERE subscription.state IN ('active', 'past_due')
          AND price.cadence IN ('monthly', 'yearly')
        ORDER BY subscription.id
    ) AS source;

    SELECT digest.row_count, digest.revision_ceiling,
           digest.canonical_sha256
    INTO projection_count_value, projection_revision_value,
         projection_sha_value
    FROM platform.billing_revenue_projection_digest_v1(p_authority_id) AS digest;

    IF source_count_value IS DISTINCT FROM projection_count_value
       OR source_revision_value IS DISTINCT FROM projection_revision_value
       OR source_sha_value IS DISTINCT FROM projection_sha_value THEN
        RAISE EXCEPTION USING ERRCODE = 'P5790',
            MESSAGE = 'billing revenue projection rebuild differs from source';
    END IF;

    INSERT INTO platform.billing_revenue_rebuilds (
        id, authority_id, actor_id, authority_group_generation,
        billing_generation, source_count, projected_count,
        source_revision_ceiling, before_sha256, source_sha256,
        after_sha256, rebuilt_at
    ) VALUES (
        p_rebuild_id, p_authority_id, p_actor_id, group_generation,
        billing_generation, source_count_value, projection_count_value,
        source_revision_value, before_sha, source_sha_value,
        projection_sha_value, p_rebuilt_at
    );

    RETURN QUERY SELECT
        source_count_value, projection_count_value,
        source_revision_value, source_sha_value, projection_sha_value;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_rebuild_revenue_projection_v1(
    uuid, uuid, uuid, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_rebuild_revenue_projection_v1(
    uuid, uuid, uuid, timestamptz
) TO harpy_runtime;
-- +goose StatementEnd
