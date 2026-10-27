-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_replace_discount_campaign_v1(
    p_authority_id uuid,
    p_operation_id uuid,
    p_discount_id uuid,
    p_old_discount_version_id uuid,
    p_new_discount_version_id uuid,
    p_campaign_id uuid,
    p_campaign_version_id_one uuid,
    p_campaign_version_id_two uuid,
    p_campaign_operation_id_one uuid,
    p_campaign_operation_id_two uuid,
    p_occurred_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_previous_campaign_id uuid;
    v_campaign_id uuid;
    v_current_id uuid;
    v_current_version bigint;
    v_current_priority integer;
    v_current_starts_at timestamptz;
    v_current_ends_at timestamptz;
    v_current_metadata jsonb;
    v_current_sha256 bytea;
    v_current_name text;
    v_current_description text;
    v_current_banner text;
    v_current_auto_apply boolean;
    v_current_available boolean;
    v_revision bigint;
    v_count smallint;
    v_position integer;
    v_version_id uuid;
    v_operation_id uuid;
BEGIN
    v_position := 0;
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    PERFORM platform.billing_assert_promotion_association_v1(
        p_authority_id, p_operation_id, p_discount_id,
        p_old_discount_version_id, p_new_discount_version_id,
        p_campaign_id, p_campaign_version_id_one, p_campaign_version_id_two,
        p_campaign_operation_id_one, p_campaign_operation_id_two, p_occurred_at
    );
    IF p_new_discount_version_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing promotion association input is invalid';
    END IF;

    IF p_old_discount_version_id IS NOT NULL THEN
        SELECT CASE WHEN pg_catalog.count(DISTINCT version_row.campaign_id) = 1
                    THEN pg_catalog.min(version_row.campaign_id::text)::uuid END
          INTO v_previous_campaign_id
        FROM platform.billing_campaign_discounts AS binding
        JOIN platform.billing_campaign_versions AS version_row
          ON version_row.authority_id = binding.authority_id
         AND version_row.id = binding.campaign_version_id
        WHERE binding.authority_id = p_authority_id
          AND binding.discount_version_id = p_old_discount_version_id
          AND NOT EXISTS (
              SELECT 1
              FROM platform.billing_campaign_versions AS successor
              WHERE successor.authority_id = version_row.authority_id
                AND successor.superseded_version_id = version_row.id
          );
        IF (
            SELECT pg_catalog.count(DISTINCT version_row.campaign_id)
            FROM platform.billing_campaign_discounts AS binding
            JOIN platform.billing_campaign_versions AS version_row
              ON version_row.authority_id = binding.authority_id
             AND version_row.id = binding.campaign_version_id
            WHERE binding.authority_id = p_authority_id
              AND binding.discount_version_id = p_old_discount_version_id
              AND NOT EXISTS (
                  SELECT 1 FROM platform.billing_campaign_versions AS successor
                  WHERE successor.authority_id = version_row.authority_id
                    AND successor.superseded_version_id = version_row.id
              )
        ) > 1 THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing discount has ambiguous campaign association';
        END IF;
    END IF;

    FOR v_campaign_id IN
        SELECT DISTINCT value
        FROM pg_catalog.unnest(ARRAY[
            v_previous_campaign_id, p_campaign_id
        ]) AS selected(value)
        WHERE value IS NOT NULL
        ORDER BY value
    LOOP
        v_position := v_position + 1;
        v_version_id := CASE v_position WHEN 1 THEN p_campaign_version_id_one
            ELSE p_campaign_version_id_two END;
        v_operation_id := CASE v_position WHEN 1 THEN p_campaign_operation_id_one
            ELSE p_campaign_operation_id_two END;
        SELECT version_row.id, version_row.version, version_row.priority,
               version_row.starts_at, version_row.ends_at,
               version_row.presentation_metadata, version_row.presentation_sha256,
               version_row.presentation_name, version_row.presentation_description,
               version_row.presentation_banner_text, version_row.auto_apply,
               version_row.available
          INTO STRICT v_current_id, v_current_version, v_current_priority,
               v_current_starts_at, v_current_ends_at, v_current_metadata,
               v_current_sha256, v_current_name, v_current_description,
               v_current_banner, v_current_auto_apply, v_current_available
        FROM platform.billing_campaign_versions AS version_row
        WHERE version_row.authority_id = p_authority_id
          AND version_row.campaign_id = v_campaign_id
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_campaign_versions AS successor
              WHERE successor.authority_id = version_row.authority_id
                AND successor.superseded_version_id = version_row.id
          )
        ORDER BY version_row.version DESC, version_row.id ASC
        LIMIT 1
        FOR SHARE;

        SELECT (
            pg_catalog.count(*) FILTER (
                WHERE discount_version.discount_id <> p_discount_id
            ) + CASE WHEN v_campaign_id = p_campaign_id THEN 1 ELSE 0 END
        )::smallint
          INTO v_count
        FROM platform.billing_campaign_discounts AS binding
        JOIN platform.billing_discount_versions AS discount_version
          ON discount_version.authority_id = binding.authority_id
         AND discount_version.id = binding.discount_version_id
        WHERE binding.authority_id = p_authority_id
          AND binding.campaign_version_id = v_current_id;

        UPDATE platform.billing_campaigns
        SET revision = revision + 1
        WHERE authority_id = p_authority_id AND id = v_campaign_id
          AND lifecycle = 'active' AND revision = v_current_version
        RETURNING revision INTO STRICT v_revision;

        INSERT INTO platform.billing_campaign_versions (
            id, authority_id, campaign_id, version, priority,
            expected_discount_count, starts_at, ends_at,
            presentation_metadata, presentation_sha256, publication_state,
            publication_operation_id, published_at, presentation_name,
            presentation_description, presentation_banner_text, auto_apply,
            available, superseded_version_id
        ) VALUES (
            v_version_id, p_authority_id, v_campaign_id, v_revision,
            v_current_priority, v_count, v_current_starts_at, v_current_ends_at,
            v_current_metadata, v_current_sha256,
            'published', v_operation_id, p_occurred_at,
            v_current_name, v_current_description,
            v_current_banner, v_current_auto_apply,
            v_current_available, v_current_id
        );

        INSERT INTO platform.billing_campaign_discounts (
            authority_id, campaign_version_id, discount_version_id, created_at
        )
        SELECT p_authority_id, v_version_id, binding.discount_version_id,
               p_occurred_at
        FROM platform.billing_campaign_discounts AS binding
        JOIN platform.billing_discount_versions AS discount_version
          ON discount_version.authority_id = binding.authority_id
         AND discount_version.id = binding.discount_version_id
        WHERE binding.authority_id = p_authority_id
          AND binding.campaign_version_id = v_current_id
          AND discount_version.discount_id <> p_discount_id;
        IF v_campaign_id = p_campaign_id THEN
            INSERT INTO platform.billing_campaign_discounts (
                authority_id, campaign_version_id, discount_version_id, created_at
            ) VALUES (
                p_authority_id, v_version_id,
                p_new_discount_version_id, p_occurred_at
            );
        END IF;
    END LOOP;

    IF v_previous_campaign_id IS DISTINCT FROM p_campaign_id THEN
        INSERT INTO platform.billing_promotion_association_receipts (
            operation_id, authority_id, discount_id,
            previous_campaign_id, campaign_id, occurred_at
        ) VALUES (
            p_operation_id, p_authority_id, p_discount_id,
            v_previous_campaign_id, p_campaign_id, p_occurred_at
        );
    END IF;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_replace_discount_campaign_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_replace_discount_campaign_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
