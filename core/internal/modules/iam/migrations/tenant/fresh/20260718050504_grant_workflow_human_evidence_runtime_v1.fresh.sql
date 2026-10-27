-- +goose Up
GRANT EXECUTE ON FUNCTION iam_authorization_evidence_read_v4(
    name,uuid,uuid,text,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_workflow_human_assignment_evidence_v1(
    name,uuid,text,uuid
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_workflow_human_candidate_evidence_v1(
    name,uuid,uuid,text,uuid
) TO forgecrud_runtime;
