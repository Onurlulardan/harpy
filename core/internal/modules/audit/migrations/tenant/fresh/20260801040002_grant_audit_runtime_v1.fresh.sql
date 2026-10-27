-- +goose Up
REVOKE ALL ON TABLE audit_entries FROM PUBLIC, forgecrud_runtime;

REVOKE ALL ON FUNCTION audit_field_changes_valid_v1(jsonb) FROM PUBLIC, forgecrud_runtime;
REVOKE ALL ON FUNCTION audit_guard_entry_immutable_v1() FROM PUBLIC, forgecrud_runtime;
REVOKE ALL ON FUNCTION audit_entry_append_v1(
    uuid,uuid,uuid,uuid,uuid,text,text,text,text,text,text,bigint,text,
    jsonb,bytea,timestamptz,timestamptz
) FROM PUBLIC, forgecrud_runtime;
REVOKE ALL ON FUNCTION audit_entries_list_v1(
    text,text,text,text,text,uuid,uuid,uuid,timestamptz,timestamptz,text,integer,integer
) FROM PUBLIC, forgecrud_runtime;
REVOKE ALL ON FUNCTION audit_entry_get_v1(uuid) FROM PUBLIC, forgecrud_runtime;
REVOKE ALL ON FUNCTION audit_stats_v1(timestamptz) FROM PUBLIC, forgecrud_runtime;

GRANT EXECUTE ON FUNCTION audit_entry_append_v1(
    uuid,uuid,uuid,uuid,uuid,text,text,text,text,text,text,bigint,text,
    jsonb,bytea,timestamptz,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION audit_entries_list_v1(
    text,text,text,text,text,uuid,uuid,uuid,timestamptz,timestamptz,text,integer,integer
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION audit_entry_get_v1(uuid) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION audit_stats_v1(timestamptz) TO forgecrud_runtime;
