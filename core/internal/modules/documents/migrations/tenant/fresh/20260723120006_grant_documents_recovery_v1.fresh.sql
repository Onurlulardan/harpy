-- +goose Up
REVOKE ALL ON FUNCTION
    documents_recovery_begin_hold_v1(uuid,uuid,uuid,bytea)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    documents_recovery_inventory_v1(uuid,uuid)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    documents_recovery_record_copy_v1(
        uuid,uuid,text,text,text,text,bytea,bigint,text
    )
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    documents_recovery_seal_hold_v1(uuid,bytea,bigint)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    documents_recovery_release_hold_v1(uuid,bytea)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    documents_recovery_fail_hold_v1(uuid,text)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    documents_recovery_apply_locator_v1(
        uuid,uuid,uuid,text,text,text,text,bigint,bytea
    )
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    documents_recovery_terminalize_transients_v1(uuid)
FROM PUBLIC;

GRANT EXECUTE ON FUNCTION
    documents_recovery_begin_hold_v1(uuid,uuid,uuid,bytea)
TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION
    documents_recovery_inventory_v1(uuid,uuid)
TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION
    documents_recovery_record_copy_v1(
        uuid,uuid,text,text,text,text,bytea,bigint,text
    )
TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION
    documents_recovery_seal_hold_v1(uuid,bytea,bigint)
TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION
    documents_recovery_release_hold_v1(uuid,bytea)
TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION
    documents_recovery_fail_hold_v1(uuid,text)
TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION
    documents_recovery_apply_locator_v1(
        uuid,uuid,uuid,text,text,text,text,bigint,bytea
    )
TO forgecrud_restore;
GRANT EXECUTE ON FUNCTION
    documents_recovery_terminalize_transients_v1(uuid)
TO forgecrud_restore;
