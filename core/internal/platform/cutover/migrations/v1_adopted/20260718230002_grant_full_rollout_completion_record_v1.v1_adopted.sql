-- +goose Up
GRANT EXECUTE ON FUNCTION cutover_authority.record_full_rollout_completion_v1(
    uuid,uuid,uuid,text,text,text,text,text,text,text,text,
    timestamptz,timestamptz,timestamptz
) TO harpy_cluster_admin;

