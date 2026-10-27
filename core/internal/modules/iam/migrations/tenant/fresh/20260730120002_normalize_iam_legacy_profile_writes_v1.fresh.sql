-- +goose Up
-- Migration stream version: 20260730120002.
-- +goose StatementBegin
CREATE FUNCTION iam_profile_display_name_write_normalize_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_display_count bigint;
    v_exact_count bigint;
    v_field_id uuid;
    v_field_key text;
    v_value_kind text;
    v_cardinality text;
    v_field_state text;
    v_canonical text;
BEGIN
    IF pg_catalog.jsonb_typeof(NEW.profile_values) = 'object'
       AND (SELECT pg_catalog.count(*)
            FROM pg_catalog.jsonb_object_keys(NEW.profile_values)) = 1
       AND NEW.profile_values ? 'display_name' THEN
        SELECT pg_catalog.count(*),
               pg_catalog.count(*) FILTER (
                   WHERE snapshot.value_kind = 'string'
                     AND snapshot.cardinality = 'single'
                     AND snapshot.field_state = 'active'
               )
        INTO v_display_count, v_exact_count
        FROM iam_profile_field_revisions AS snapshot
        WHERE snapshot.schema_version_id = NEW.profile_schema_version_id
          AND snapshot.resource_kind = NEW.profile_resource_kind
          AND snapshot.field_key = 'display_name';
        IF v_display_count IS DISTINCT FROM 1
           OR v_exact_count IS DISTINCT FROM 1 THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'IAM profile display name snapshot is invalid';
        END IF;
        SELECT snapshot.field_id, snapshot.field_key, snapshot.value_kind,
               snapshot.cardinality, snapshot.field_state
        INTO STRICT v_field_id, v_field_key, v_value_kind,
                    v_cardinality, v_field_state
        FROM iam_profile_field_revisions AS snapshot
        WHERE snapshot.schema_version_id = NEW.profile_schema_version_id
          AND snapshot.resource_kind = NEW.profile_resource_kind
          AND snapshot.field_key = 'display_name';
        v_canonical := iam_profile_installation_legacy_document_v1(
            NEW.profile_values, NEW.profile_sha256, v_field_id,
            v_field_key, v_value_kind, v_cardinality, v_field_state
        );
        NEW.profile_values := v_canonical::jsonb;
        NEW.profile_sha256 := pg_catalog.sha256(
            pg_catalog.convert_to(v_canonical, 'UTF8')
        );
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_display_name_write_normalize_v1()
    FROM PUBLIC;

CREATE TRIGGER iam_users_profile_display_name_write_normalize
BEFORE INSERT OR UPDATE OF profile_values,profile_sha256 ON iam_users
FOR EACH ROW EXECUTE FUNCTION iam_profile_display_name_write_normalize_v1();
ALTER TABLE iam_users ENABLE ALWAYS TRIGGER
    iam_users_profile_display_name_write_normalize;

CREATE TRIGGER iam_organizations_profile_display_name_write_normalize
BEFORE INSERT OR UPDATE OF profile_values,profile_sha256 ON iam_organizations
FOR EACH ROW EXECUTE FUNCTION iam_profile_display_name_write_normalize_v1();
ALTER TABLE iam_organizations ENABLE ALWAYS TRIGGER
    iam_organizations_profile_display_name_write_normalize;

CREATE TRIGGER iam_roles_profile_display_name_write_normalize
BEFORE INSERT OR UPDATE OF profile_values,profile_sha256 ON iam_roles
FOR EACH ROW EXECUTE FUNCTION iam_profile_display_name_write_normalize_v1();
ALTER TABLE iam_roles ENABLE ALWAYS TRIGGER
    iam_roles_profile_display_name_write_normalize;
