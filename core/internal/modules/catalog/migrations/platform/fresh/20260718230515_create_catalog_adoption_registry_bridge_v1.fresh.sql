-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.catalog_adoption_registry_blocker_v1(
    p_package_id uuid,p_slug text,p_version text,p_manifest_sha256 text
) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_package_id uuid;
    v_manifest_sha256 text;
BEGIN
    IF SESSION_USER NOT IN ('forgecrud_v2_runtime','forgecrud_adopter') OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog adoption registry authority denied';
    END IF;
    SELECT package.package_id INTO v_package_id
    FROM platform.catalog_packages AS package
    WHERE package.slug=p_slug
      AND (p_package_id IS NULL OR package.package_id=p_package_id);
    IF NOT FOUND THEN
        RETURN 'UNKNOWN_PACKAGE';
    END IF;
    SELECT version.manifest_sha256 INTO v_manifest_sha256
    FROM platform.catalog_package_versions AS version
    WHERE version.package_id=v_package_id AND version.version=p_version;
    IF NOT FOUND THEN
        RETURN 'UNKNOWN_VERSION';
    END IF;
    IF p_manifest_sha256 IS NOT NULL AND
       p_manifest_sha256 IS DISTINCT FROM v_manifest_sha256 THEN
        RETURN 'MANIFEST_HASH_MISMATCH';
    END IF;
    RETURN NULL;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.catalog_adoption_registry_blocker_v1(
    uuid,text,text,text
) FROM PUBLIC;
