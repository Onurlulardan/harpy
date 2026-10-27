SELECT assignment_document
FROM iam_workflow_human_assignment_evidence_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,
    $2::text,
    $3::uuid
);
