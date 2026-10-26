SELECT run.run_id, run.source_count, run.target_count, run.source_sha256, run.target_sha256,
       run.state = 'ACCEPTED',
       fence.state IN ('V2_FENCED','RETIRED') AND fence.run_id = run.run_id,
       (SELECT count(*) FROM work_audit_entries AS audit
        WHERE audit.occurred_at > run.created_at
          AND audit.action NOT LIKE 'work.v1_adoption.%'),
       (SELECT COALESCE(sum(hit_count), 0) FROM work_compatibility_hits
        WHERE retired_at IS NULL)
FROM work_adoption_runs AS run
CROSS JOIN work_v1_writer_fence AS fence
WHERE run.run_id = $1 AND fence.singleton
