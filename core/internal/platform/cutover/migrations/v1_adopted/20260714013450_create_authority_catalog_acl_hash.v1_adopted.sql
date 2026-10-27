-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.catalog_acl_sha256_v1(p_kind text, p_object_oid oid)
RETURNS text LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_located_count integer;
BEGIN
    IF p_kind IS NULL OR p_kind NOT IN ('schema','table','function') OR
       p_object_oid IS NULL OR p_object_oid = 0 THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'catalog ACL identity is invalid';
    END IF;
    SELECT pg_catalog.count(*) INTO v_located_count FROM (
        SELECT namespace.oid FROM pg_catalog.pg_namespace AS namespace
        WHERE p_kind = 'schema' AND namespace.oid = p_object_oid
        UNION ALL
        SELECT relation.oid FROM pg_catalog.pg_class AS relation
        WHERE p_kind = 'table' AND relation.oid = p_object_oid
          AND relation.relkind IN ('r','p')
        UNION ALL
        SELECT procedure_row.oid FROM pg_catalog.pg_proc AS procedure_row
        WHERE p_kind = 'function' AND procedure_row.oid = p_object_oid
    ) AS located;
    IF v_located_count IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'catalog ACL object is absent';
    END IF;
RETURN (SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
    pg_catalog.jsonb_build_object(
        'namespace', COALESCE((SELECT pg_catalog.jsonb_agg(
            pg_catalog.jsonb_build_object(
                'grantor', pg_catalog.pg_get_userbyid(entry.grantor),
                'grantee', CASE WHEN entry.grantee = 0 THEN 'PUBLIC'
                    ELSE pg_catalog.pg_get_userbyid(entry.grantee) END,
                'privilege', entry.privilege_type,
                'grantable', entry.is_grantable)
            ORDER BY CASE WHEN entry.grantee = 0 THEN 'PUBLIC'
                         ELSE pg_catalog.pg_get_userbyid(entry.grantee) END,
                     pg_catalog.pg_get_userbyid(entry.grantor),
                     entry.privilege_type, entry.is_grantable)
            FROM pg_catalog.pg_namespace AS namespace
            CROSS JOIN pg_catalog.aclexplode(COALESCE(namespace.nspacl,
                pg_catalog.acldefault('n', namespace.nspowner))) AS entry
            WHERE p_kind = 'schema' AND namespace.oid = p_object_oid), '[]'::jsonb),
        'relation', COALESCE((SELECT pg_catalog.jsonb_build_object(
            'table', COALESCE((SELECT pg_catalog.jsonb_agg(
                pg_catalog.jsonb_build_object(
                    'grantor', pg_catalog.pg_get_userbyid(entry.grantor),
                    'grantee', CASE WHEN entry.grantee = 0 THEN 'PUBLIC'
                        ELSE pg_catalog.pg_get_userbyid(entry.grantee) END,
                    'privilege', entry.privilege_type, 'grantable', entry.is_grantable)
                ORDER BY CASE WHEN entry.grantee = 0 THEN 'PUBLIC'
                             ELSE pg_catalog.pg_get_userbyid(entry.grantee) END,
                         pg_catalog.pg_get_userbyid(entry.grantor),
                         entry.privilege_type, entry.is_grantable)
                FROM pg_catalog.aclexplode(COALESCE(relation.relacl,
                    pg_catalog.acldefault('r'::"char", relation.relowner))) AS entry), '[]'::jsonb),
            'columns', COALESCE((SELECT pg_catalog.jsonb_agg(
                pg_catalog.jsonb_build_object(
                    'number', attribute.attnum, 'name', attribute.attname,
                    'acl', COALESCE((SELECT pg_catalog.jsonb_agg(
                        pg_catalog.jsonb_build_object(
                            'grantor', pg_catalog.pg_get_userbyid(column_entry.grantor),
                            'grantee', CASE WHEN column_entry.grantee = 0 THEN 'PUBLIC'
                                ELSE pg_catalog.pg_get_userbyid(column_entry.grantee) END,
                            'privilege', column_entry.privilege_type,
                            'grantable', column_entry.is_grantable)
                        ORDER BY CASE WHEN column_entry.grantee = 0 THEN 'PUBLIC'
                                     ELSE pg_catalog.pg_get_userbyid(column_entry.grantee) END,
                            pg_catalog.pg_get_userbyid(column_entry.grantor),
                            column_entry.privilege_type, column_entry.is_grantable)
                        FROM pg_catalog.aclexplode(attribute.attacl) AS column_entry), '[]'::jsonb))
                ORDER BY attribute.attnum) FROM pg_catalog.pg_attribute AS attribute
                WHERE attribute.attrelid = relation.oid AND attribute.attnum > 0
                  AND NOT attribute.attisdropped), '[]'::jsonb))
            FROM pg_catalog.pg_class AS relation
            WHERE p_kind = 'table' AND relation.oid = p_object_oid), '{}'::jsonb),
        'routine', COALESCE((SELECT pg_catalog.jsonb_agg(
            pg_catalog.jsonb_build_object(
                'grantor', pg_catalog.pg_get_userbyid(entry.grantor),
                'grantee', CASE WHEN entry.grantee = 0 THEN 'PUBLIC'
                    ELSE pg_catalog.pg_get_userbyid(entry.grantee) END,
                'privilege', entry.privilege_type,
                'grantable', entry.is_grantable)
            ORDER BY CASE WHEN entry.grantee = 0 THEN 'PUBLIC'
                         ELSE pg_catalog.pg_get_userbyid(entry.grantee) END,
                     pg_catalog.pg_get_userbyid(entry.grantor),
                     entry.privilege_type, entry.is_grantable)
            FROM pg_catalog.pg_proc AS procedure_row
            CROSS JOIN pg_catalog.aclexplode(COALESCE(procedure_row.proacl,
                pg_catalog.acldefault('f', procedure_row.proowner))) AS entry
            WHERE p_kind = 'function' AND procedure_row.oid = p_object_oid), '[]'::jsonb)
    )::text, 'UTF8')), 'hex'));
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.catalog_acl_sha256_v1(text, oid) FROM PUBLIC;
