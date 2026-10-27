-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_locator_source_sha256_v1(
    p_token_id uuid,
    p_family_id uuid,
    p_selector bytea,
    p_revision bigint,
    p_state text,
    p_secret_hmac bytea
)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF p_token_id IS NOT DISTINCT FROM
           '00000000-0000-0000-0000-000000000000'::uuid OR
       p_family_id IS NOT DISTINCT FROM
           '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       p_revision OPERATOR(pg_catalog.<=) 0 OR
       p_state NOT IN ('active', 'consumed', 'revoked', 'expired') OR
       pg_catalog.octet_length(p_secret_hmac) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM refresh locator source input is invalid';
    END IF;
    RETURN pg_catalog.sha256(
        pg_catalog.convert_to(
            'forgecrud:iam-refresh-locator-source:v1', 'UTF8'
        ) OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
          OPERATOR(pg_catalog.||) pg_catalog.uuid_send(p_token_id)
          OPERATOR(pg_catalog.||) pg_catalog.uuid_send(p_family_id)
          OPERATOR(pg_catalog.||) p_selector
          OPERATOR(pg_catalog.||) pg_catalog.int8send(p_revision)
          OPERATOR(pg_catalog.||) pg_catalog.convert_to(p_state, 'UTF8')
          OPERATOR(pg_catalog.||) pg_catalog.decode('00', 'hex')
          OPERATOR(pg_catalog.||) p_secret_hmac
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_refresh_locator_source_sha256_v1(
    uuid,uuid,bytea,bigint,text,bytea
) FROM PUBLIC;
