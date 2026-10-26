SELECT evidence_document
FROM iam_workflow_human_candidate_evidence_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,
    $2::uuid,
    $3::text,
    $4::uuid
);
