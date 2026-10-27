-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_add_discount_offering_v1(
    p_authority_id uuid, p_discount_version_id uuid,
    p_offering_version_id uuid, p_occurred_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    INSERT INTO platform.billing_discount_offerings (
        authority_id, discount_version_id, offering_version_id, created_at
    )
    SELECT p_authority_id, p_discount_version_id,
           p_offering_version_id, p_occurred_at
    FROM platform.billing_discount_versions AS discount_version
    JOIN platform.billing_offering_versions AS offering_version
      ON offering_version.authority_id = discount_version.authority_id
     AND offering_version.id = p_offering_version_id
    WHERE discount_version.authority_id = p_authority_id
      AND discount_version.id = p_discount_version_id
      AND NOT discount_version.all_offerings;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '23503',
            MESSAGE = 'billing discount offering binding is invalid';
    END IF;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_add_discount_offering_v1(
    uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_add_discount_offering_v1(
    uuid,uuid,uuid,timestamptz
) TO harpy_runtime;
-- +goose StatementEnd

