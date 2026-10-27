-- +goose Up
GRANT USAGE ON SCHEMA platform TO forgecrud_backup_reader;
GRANT USAGE ON SCHEMA platform TO forgecrud_restore;
GRANT USAGE ON SCHEMA platform TO forgecrud_runtime;

REVOKE ALL ON FUNCTION platform.platformops_record_signing_bundle_v1(uuid,bigint,text,bytea,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_get_signing_bundle_v1(uuid,bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_request_recovery_set_v1(uuid,uuid,uuid,uuid,text,uuid,bigint,uuid,bigint,text,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_claim_recovery_set_v1(uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_advance_recovery_set_v1(uuid,text,text,text,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_record_recovery_component_v1(uuid,text,text,text,text,bigint,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_complete_recovery_set_v1(uuid,text,bytea,uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_get_recovery_set_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_list_retention_candidates_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_mark_recovery_set_deleting_v1(uuid,text,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_finish_recovery_set_delete_v1(uuid,boolean,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_request_restore_v1(uuid,uuid,uuid,text,text,bytea,uuid,text,text,text,boolean,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_claim_restore_v1(uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_advance_restore_v1(uuid,text,text,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_complete_restore_v1(uuid,text,bytea,uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_get_restore_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_record_signing_bundle_v1(uuid,bigint,text,bytea,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_get_signing_bundle_v1(uuid,bigint) TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_get_signing_bundle_v1(uuid,bigint) TO forgecrud_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_get_signing_bundle_v1(uuid,bigint) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_get_recovery_set_v1(uuid) TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_get_recovery_set_v1(uuid) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_list_retention_candidates_v1(uuid) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_mark_recovery_set_deleting_v1(uuid,text,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_finish_recovery_set_delete_v1(uuid,boolean,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_request_restore_v1(uuid,uuid,uuid,text,text,bytea,uuid,text,text,text,boolean,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_get_restore_v1(uuid) TO forgecrud_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_get_restore_v1(uuid) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION platform.platformops_claim_recovery_set_v1(uuid,timestamptz) TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_advance_recovery_set_v1(uuid,text,text,text,timestamptz) TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_record_recovery_component_v1(uuid,text,text,text,text,bigint,timestamptz) TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_complete_recovery_set_v1(uuid,text,bytea,uuid,timestamptz) TO forgecrud_backup_reader;

GRANT EXECUTE ON FUNCTION platform.platformops_claim_restore_v1(uuid,timestamptz) TO forgecrud_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_advance_restore_v1(uuid,text,text,timestamptz) TO forgecrud_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_complete_restore_v1(uuid,text,bytea,uuid,timestamptz) TO forgecrud_restore;
