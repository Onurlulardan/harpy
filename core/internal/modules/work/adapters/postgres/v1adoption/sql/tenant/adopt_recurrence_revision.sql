INSERT INTO work_recurrence_rule_revisions (
    rule_id, revision, contract_version, dtstart, timezone_name, tzdb_version,
    rrule, exdates, effective_from, created_at
) VALUES ($1,1,'work_rrule_v1',$2,$3,$4,$5,ARRAY[]::text[],$6,transaction_timestamp())
ON CONFLICT (rule_id, revision) DO NOTHING
