-- +goose Up
REVOKE ALL ON FUNCTION platform.platformops_record_trace_v1(uuid,uuid,text,text,text,text,text,text,text,text,integer,timestamptz,timestamptz,numeric,text,text,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_list_traces_v1(uuid,text,text,text,integer,integer,integer,numeric,numeric,timestamptz,timestamptz,text,text,text,integer,integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_get_trace_v1(uuid,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_trace_stats_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_prune_observability_v1(uuid,timestamptz,integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_list_alert_rules_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_create_alert_rule_v1(uuid,uuid,text,text,boolean,text,text,integer,jsonb,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_update_alert_rule_v1(uuid,uuid,text,text,boolean,text,text,integer,jsonb,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_delete_alert_rule_v1(uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_test_alert_rule_v1(uuid,uuid,uuid,uuid,text,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_list_alert_events_v1(uuid,text,text,integer,integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_record_trace_v1(uuid,uuid,text,text,text,text,text,text,text,text,integer,timestamptz,timestamptz,numeric,text,text,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_list_traces_v1(uuid,text,text,text,integer,integer,integer,numeric,numeric,timestamptz,timestamptz,text,text,text,integer,integer) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_get_trace_v1(uuid,text) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_trace_stats_v1(uuid) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_prune_observability_v1(uuid,timestamptz,integer) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_list_alert_rules_v1(uuid) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_create_alert_rule_v1(uuid,uuid,text,text,boolean,text,text,integer,jsonb,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_update_alert_rule_v1(uuid,uuid,text,text,boolean,text,text,integer,jsonb,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_delete_alert_rule_v1(uuid,uuid) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_test_alert_rule_v1(uuid,uuid,uuid,uuid,text,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_list_alert_events_v1(uuid,text,text,integer,integer) TO forgecrud_runtime;
