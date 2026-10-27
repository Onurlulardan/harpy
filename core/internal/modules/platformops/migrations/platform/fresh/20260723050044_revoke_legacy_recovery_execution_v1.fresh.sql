-- +goose Up
REVOKE EXECUTE ON FUNCTION
    platform.platformops_claim_recovery_set_v1(uuid,timestamptz)
FROM forgecrud_backup_reader;
REVOKE EXECUTE ON FUNCTION
    platform.platformops_advance_recovery_set_v1(uuid,text,text,text,timestamptz)
FROM forgecrud_backup_reader;
REVOKE EXECUTE ON FUNCTION
    platform.platformops_record_recovery_component_v1(
        uuid,text,text,text,text,bigint,timestamptz
    )
FROM forgecrud_backup_reader;
REVOKE EXECUTE ON FUNCTION
    platform.platformops_complete_recovery_set_v1(
        uuid,text,bytea,uuid,timestamptz
    )
FROM forgecrud_backup_reader;
