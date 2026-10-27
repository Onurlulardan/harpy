-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_installation_derived_id_v1(
    p_operation_id uuid, p_label text
)
RETURNS uuid
LANGUAGE plpgsql
IMMUTABLE
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_digest bytea;
    v_hex text;
BEGIN
    IF p_operation_id IS NULL OR p_label IS NULL
       OR pg_catalog.octet_length(p_label) NOT BETWEEN 1 AND 64
       OR p_label !~ '^[a-z0-9-]+$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM installation identity derivation input is invalid';
    END IF;
    v_digest := pg_catalog.sha256(
        pg_catalog.convert_to('forgecrud:iam:installation:v1', 'UTF8') ||
        pg_catalog.uuid_send(p_operation_id) || '\x00'::bytea ||
        pg_catalog.convert_to(p_label, 'UTF8')
    );
    v_digest := pg_catalog.set_byte(
        v_digest, 6, (pg_catalog.get_byte(v_digest, 6) & 15) | 112
    );
    v_digest := pg_catalog.set_byte(
        v_digest, 8, (pg_catalog.get_byte(v_digest, 8) & 63) | 128
    );
    v_hex := pg_catalog.encode(pg_catalog.substring(v_digest, 1, 16), 'hex');
    RETURN (
        pg_catalog.substring(v_hex, 1, 8) || '-' ||
        pg_catalog.substring(v_hex, 9, 4) || '-' ||
        pg_catalog.substring(v_hex, 13, 4) || '-' ||
        pg_catalog.substring(v_hex, 17, 4) || '-' ||
        pg_catalog.substring(v_hex, 21, 12)
    )::uuid;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_derived_id_v1(uuid,text) FROM PUBLIC;
