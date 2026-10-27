-- +goose Up
-- +goose StatementBegin
GRANT USAGE ON SCHEMA platform TO harpy_runtime, harpy_adopter;

GRANT EXECUTE ON FUNCTION platform.usage_record_fact_v1(
uuid,uuid,uuid,uuid,bytea,bytea,text,text,integer,bigint,timestamptz) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.usage_active_users_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,timestamptz)
TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.usage_top_users_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,
timestamptz,timestamptz,integer) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.usage_user_summary_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,text,
timestamptz,timestamptz) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.usage_user_timeline_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,text,
integer,integer) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.usage_features_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,
timestamptz,timestamptz,integer) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.usage_heatmap_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,
timestamptz,timestamptz) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.usage_trend_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,
timestamptz,timestamptz,text) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.usage_overview_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,
timestamptz,timestamptz,timestamptz) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.usage_legal_hold_place_v1(
uuid,uuid,uuid,text,timestamptz,timestamptz,timestamptz) TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.usage_legal_hold_release_v1(uuid,timestamptz)
TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.usage_retention_prune_v1(
uuid,uuid,timestamptz,integer,timestamptz) TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.usage_adoption_start_v1(
uuid,uuid,text,text,text,oid,bigint,bytea,text,bytea,timestamptz) TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.usage_adoption_claim_v1(
uuid,uuid,uuid,bigint,oid,bigint,bytea,timestamptz) TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.usage_adoption_apply_page_v1(
uuid,uuid,uuid,bigint,bigint,uuid,uuid,jsonb,bytea,bytea,oid,bigint,bytea,timestamptz)
TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.usage_adoption_advance_v1(
uuid,uuid,uuid,bigint,bigint,uuid,uuid,bytea,bytea,oid,bigint,bytea,timestamptz)
TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.usage_adoption_reconcile_v1(
uuid,uuid,uuid,bigint,oid,bigint,bytea,timestamptz) TO harpy_adopter;
GRANT EXECUTE ON FUNCTION platform.usage_adoption_complete_v1(
uuid,uuid,uuid,bigint,oid,bigint,bytea,timestamptz) TO harpy_adopter;
-- +goose StatementEnd
