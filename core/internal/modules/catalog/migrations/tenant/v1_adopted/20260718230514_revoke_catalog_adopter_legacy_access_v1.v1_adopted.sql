-- +goose Up
-- Direct relation privileges are never assigned to harpy_adopter.
-- Revoke every internal routine so the composite stays its sole surface.
REVOKE EXECUTE ON FUNCTION catalog_catalog_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_catalog_component_blockers_v1(
    uuid,uuid,uuid,bigint,bigint,timestamptz
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_delete_dependency_resolutions_v1(uuid)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_adoption_inventory_v1(uuid) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_adoption_inventory_root_v1(uuid,text)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_adoption_blockers_v1(uuid) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_adoption_blocker_sha256_v1(uuid)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_v1_adoption_receipt_immutable_v1()
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_v1_adoption_value_valid_v1(jsonb,jsonb)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_v1_adoption_batch_decode_v1(bytea,text,bigint)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_v1_adoption_batch_assert_v1(
    jsonb,uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,pg_lsn,
    pg_lsn,xid8,integer,text,jsonb
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_v1_adoption_apply_v1(
    uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_v1_adoption_project_core_v1(text,jsonb)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_v1_adoption_project_menu_v1(text,jsonb)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_v1_adoption_activate_menu_v1(jsonb)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_v1_adoption_consumer_sha256_v1(uuid)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_v1_adoption_projection_ready_v1(uuid,uuid,uuid)
FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION catalog_v1_adoption_finalize_v1(uuid,uuid,uuid)
FROM harpy_adopter;
