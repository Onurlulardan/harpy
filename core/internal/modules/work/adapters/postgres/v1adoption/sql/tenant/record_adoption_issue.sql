INSERT INTO work_adoption_issues (
    run_id, issue_id, source_kind, source_id, issue_code, evidence_sha256, created_at
) VALUES ($1, $2, $3, $4, $5, $6, transaction_timestamp())
ON CONFLICT (run_id, issue_id) DO NOTHING
