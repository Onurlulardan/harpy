-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_security_epoch_initialize_v1(
    p_environment_id uuid,
    p_now timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
       OR p_environment_id IS NULL OR p_now IS NULL
       OR NOT pg_catalog.isfinite(p_now) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM security epoch initialization denied';
    END IF;
    INSERT INTO platform.iam_security_epochs (
        environment_id, security_epoch, generation, created_at, rotated_at
    ) VALUES (
        p_environment_id,
        pg_catalog.sha256(
            pg_catalog.uuid_send(pg_catalog.gen_random_uuid()) ||
            pg_catalog.uuid_send(pg_catalog.gen_random_uuid())
        ),
        1, p_now, p_now
    ) ON CONFLICT (environment_id) DO NOTHING;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_security_epoch_initialize_v1(
    uuid,timestamptz
) FROM PUBLIC;
