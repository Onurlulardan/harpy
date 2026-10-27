-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_organization_owners_replace_v1(
    p_organization_id uuid, p_expected_revision bigint,
    p_owner_ids uuid[], p_membership_ids uuid[]
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_status text; v_revision bigint; v_user_id uuid;
    v_changed boolean := false;
BEGIN
    IF p_organization_id IS NULL OR p_expected_revision < 1 OR
       COALESCE(pg_catalog.cardinality(p_owner_ids), -1) NOT BETWEEN 1 AND 100 OR
       p_membership_ids IS NULL OR pg_catalog.cardinality(p_owner_ids) <>
          pg_catalog.cardinality(p_membership_ids) OR
       p_owner_ids IS DISTINCT FROM ARRAY(
          SELECT value FROM pg_catalog.unnest(p_owner_ids) AS desired(value)
          ORDER BY value
       ) OR pg_catalog.cardinality(p_owner_ids) <> (
          SELECT pg_catalog.count(DISTINCT value)
          FROM pg_catalog.unnest(p_owner_ids) AS desired(value)
       ) OR pg_catalog.cardinality(p_membership_ids) <> (
          SELECT pg_catalog.count(DISTINCT value)
          FROM pg_catalog.unnest(p_membership_ids) AS desired(value)
       ) OR EXISTS (
          SELECT 1 FROM pg_catalog.generate_subscripts(p_owner_ids, 1) AS desired(position)
          WHERE p_owner_ids[desired.position] IS NULL OR
                p_membership_ids[desired.position] IS NULL
       ) THEN RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE =
          'IAM identity organization owner set is not canonical'; END IF;
    PERFORM membership.id FROM iam_organization_memberships AS membership
    WHERE membership.organization_id = p_organization_id
    ORDER BY membership.user_id, membership.id FOR UPDATE;
    SELECT organization.status, organization.revision INTO v_status, v_revision
    FROM iam_organizations AS organization
    WHERE organization.id = p_organization_id FOR UPDATE;
    IF NOT FOUND OR v_status <> 'active' OR v_revision <> p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE =
            'IAM identity organization owner compare-and-swap failed';
    END IF;
    IF (SELECT pg_catalog.count(*) FROM iam_users AS principal
        WHERE principal.id = ANY(p_owner_ids) AND principal.status = 'active') <>
       pg_catalog.cardinality(p_owner_ids) OR EXISTS (
        SELECT 1 FROM iam_organization_memberships AS membership
        WHERE membership.organization_id = p_organization_id
          AND membership.user_id = ANY(p_owner_ids)
          AND membership.state = 'tombstoned'
       ) OR EXISTS (
        SELECT 1
        FROM pg_catalog.generate_subscripts(p_owner_ids, 1) AS desired(position)
        JOIN iam_organization_memberships AS collision
          ON collision.id = p_membership_ids[desired.position]
        WHERE NOT EXISTS (
          SELECT 1 FROM iam_organization_memberships AS relation
          WHERE relation.organization_id = p_organization_id
            AND relation.user_id = p_owner_ids[desired.position]
        )
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE =
          'IAM identity organization owner set is invalid'; END IF;
    FOR v_user_id IN
      WITH demoted AS (
        UPDATE iam_organization_memberships AS membership
        SET membership_kind = 'member', revision = membership.revision + 1,
            updated_at = v_now
        WHERE membership.organization_id = p_organization_id
          AND membership.membership_kind = 'owner'
          AND membership.state <> 'tombstoned'
          AND NOT membership.user_id = ANY(p_owner_ids)
        RETURNING membership.user_id
    ), promoted AS (
        UPDATE iam_organization_memberships AS membership
        SET membership_kind = 'owner', state = 'active',
            revision = membership.revision + 1, updated_at = v_now
        WHERE membership.organization_id = p_organization_id
          AND membership.user_id = ANY(p_owner_ids)
          AND membership.state <> 'tombstoned'
          AND (membership.membership_kind <> 'owner' OR membership.state <> 'active')
        RETURNING membership.user_id
    ), inserted AS (
        INSERT INTO iam_organization_memberships (
            id, organization_id, user_id, membership_kind, state,
            revision, created_at, updated_at
        ) SELECT p_membership_ids[desired.position], p_organization_id,
            p_owner_ids[desired.position], 'owner', 'active', 1, v_now, v_now
        FROM pg_catalog.generate_subscripts(p_owner_ids, 1) AS desired(position)
        WHERE NOT EXISTS (
            SELECT 1 FROM iam_organization_memberships AS membership
            WHERE membership.organization_id = p_organization_id
              AND membership.user_id = p_owner_ids[desired.position]
        ) RETURNING user_id
    ) SELECT changed.user_id FROM (
        SELECT user_id FROM demoted UNION SELECT user_id FROM promoted
        UNION SELECT user_id FROM inserted
      ) AS changed ORDER BY changed.user_id
    LOOP
        v_changed := true;
        PERFORM iam_identity_user_bump_v1(v_user_id, false);
    END LOOP;
    IF (SELECT pg_catalog.count(*) FROM iam_organization_memberships AS membership
        JOIN iam_users AS principal ON principal.id = membership.user_id
        WHERE membership.organization_id = p_organization_id
          AND membership.membership_kind = 'owner' AND membership.state = 'active'
          AND principal.status = 'active') <> pg_catalog.cardinality(p_owner_ids) OR
       EXISTS (
        SELECT 1 FROM iam_organization_memberships AS membership
        WHERE membership.organization_id = p_organization_id
          AND membership.membership_kind = 'owner' AND membership.state = 'active'
          AND NOT membership.user_id = ANY(p_owner_ids)
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE =
          'IAM identity organization requires the exact active owner set'; END IF;
    RETURN v_changed;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_organization_owners_replace_v1(
    uuid,bigint,uuid[],uuid[]
) FROM PUBLIC;
