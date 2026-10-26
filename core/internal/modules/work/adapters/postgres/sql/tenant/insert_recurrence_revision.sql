INSERT INTO work_recurrence_rule_revisions (
    rule_id, revision, contract_version, dtstart, timezone_name, tzdb_version,
    rrule, exdates, effective_from, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, transaction_timestamp(), transaction_timestamp())
