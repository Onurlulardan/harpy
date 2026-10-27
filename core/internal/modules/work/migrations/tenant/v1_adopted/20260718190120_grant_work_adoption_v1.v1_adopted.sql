-- +goose Up
GRANT SELECT ON work_v1_writer_fence TO harpy_runtime;
GRANT UPDATE (state, run_id, generation, reverse_reconciliation_id, updated_by, updated_at)
ON work_v1_writer_fence TO harpy_runtime;

GRANT INSERT (run_id, cutover_generation, manifest_sha256, transform_version, state,
    source_count, target_count, source_sha256, target_sha256, final_lsn, applied_lsn,
    acceptance_sha256, created_at, updated_at)
ON work_adoption_runs TO harpy_runtime;
GRANT UPDATE (state, source_count, target_count, source_sha256, target_sha256,
    final_lsn, applied_lsn, acceptance_sha256, updated_at)
ON work_adoption_runs TO harpy_runtime;
GRANT INSERT (run_id, source_kind, source_id, target_kind, target_id, source_sha256,
    target_sha256, applied_lsn, state, issue_code, updated_at)
ON work_adoption_objects TO harpy_runtime;
GRANT UPDATE (source_sha256, target_sha256, applied_lsn, state, issue_code, updated_at)
ON work_adoption_objects TO harpy_runtime;
GRANT INSERT (run_id, issue_id, source_kind, source_id, issue_code, evidence_sha256, created_at)
ON work_adoption_issues TO harpy_runtime;
GRANT INSERT (route_contract, consumer_id, build_id, hit_count, first_hit_at, last_hit_at, retired_at)
ON work_compatibility_hits TO harpy_runtime;
GRANT UPDATE (hit_count, last_hit_at, retired_at)
ON work_compatibility_hits TO harpy_runtime;
GRANT INSERT (reconciliation_id, source_generation, target_generation, source_count,
    target_count, source_sha256, target_sha256, state, requested_by, approved_by,
    created_at, decided_at)
ON work_reverse_reconciliations TO harpy_runtime;
GRANT UPDATE (state, approved_by, decided_at) ON work_reverse_reconciliations TO harpy_runtime;
