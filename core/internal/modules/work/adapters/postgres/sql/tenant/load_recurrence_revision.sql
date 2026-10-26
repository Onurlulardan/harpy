SELECT revision, contract_version, dtstart, timezone_name, tzdb_version, rrule, exdates
FROM work_recurrence_rule_revisions revision
JOIN work_recurrence_rules rule ON rule.id = revision.rule_id
    AND rule.current_revision = revision.revision
WHERE rule.id = $1 AND rule.lifecycle_state = 'ACTIVE'
FOR UPDATE OF rule
