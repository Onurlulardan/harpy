SELECT run.id, run.lifecycle_state, checkpoint.last_source_id,
       COALESCE(checkpoint.adopted_count, 0), checkpoint.source_checksum,
       checkpoint.target_checksum,
       (SELECT pg_catalog.count(*) FROM notifications_adoption_issues issue
        WHERE issue.run_id = run.id AND issue.resolved_at IS NULL)
FROM notifications_adoption_runs run
LEFT JOIN notifications_adoption_checkpoints checkpoint
  ON checkpoint.run_id = run.id AND checkpoint.source_table = 'notifications'
JOIN notifications_writer_authority authority
  ON authority.singleton AND authority.writer = 'v1'
 AND authority.lifecycle_state = 'writable'
WHERE run.source_generation = $1 AND run.source_manifest_digest = $2::bytea
FOR UPDATE OF run;
