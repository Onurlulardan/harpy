-- +goose Up
GRANT SELECT ON
    work_boards, work_board_columns, work_board_memberships, work_tasks,
    work_task_assignees, work_labels, work_task_labels, work_checklist_items,
    work_comments, work_recurrence_rules, work_recurrence_rule_revisions,
    work_calendar_events, work_resource_links, work_inbound_references,
    work_projection_streams, work_projection_receipts, work_inbox_items,
    work_calendar_items, work_operations, work_aggregate_sequences,
    work_audit_entries, work_purge_operations, work_purge_steps,
    work_id_tombstones, work_adoption_runs, work_adoption_objects,
    work_adoption_issues, work_compatibility_hits, work_reverse_reconciliations
TO harpy_runtime;

GRANT INSERT (id, creator_user_id, name, description, lifecycle_state, reference_epoch, row_version, created_at, updated_at, archived_at)
ON work_boards TO harpy_runtime;
GRANT UPDATE (name, description, lifecycle_state, row_version, updated_at, archived_at)
ON work_boards TO harpy_runtime;
GRANT INSERT (id, board_id, name, status_category, position, row_version, created_at, updated_at)
ON work_board_columns TO harpy_runtime;
GRANT UPDATE (name, status_category, position, row_version, updated_at)
ON work_board_columns TO harpy_runtime;
GRANT INSERT (id, board_id, principal_kind, principal_id, board_role, active, row_version, created_at, updated_at, deactivated_at)
ON work_board_memberships TO harpy_runtime;
GRANT UPDATE (principal_kind, principal_id, board_role, active, row_version, updated_at, deactivated_at)
ON work_board_memberships TO harpy_runtime;
GRANT INSERT (id, board_id, column_id, task_kind, status_category, title, description, position, row_version, due_at, completed_at, cancelled_at, deleted_at, created_at, updated_at)
ON work_tasks TO harpy_runtime;
GRANT UPDATE (column_id, status_category, title, description, position, row_version, due_at, completed_at, cancelled_at, deleted_at, updated_at)
ON work_tasks TO harpy_runtime;
GRANT INSERT (task_id, principal_kind, principal_id, active, removed_at, created_at, updated_at)
ON work_task_assignees TO harpy_runtime;
GRANT UPDATE (active, removed_at, updated_at) ON work_task_assignees TO harpy_runtime;
GRANT INSERT (id, board_id, name, color, row_version, deleted_at, created_at, updated_at) ON work_labels TO harpy_runtime;
GRANT UPDATE (name, color, row_version, deleted_at, updated_at) ON work_labels TO harpy_runtime;
GRANT INSERT (task_id, label_id, active, removed_at, created_at, updated_at)
ON work_task_labels TO harpy_runtime;
GRANT UPDATE (active, removed_at, updated_at) ON work_task_labels TO harpy_runtime;
GRANT INSERT (id, task_id, item_text, completed, position, row_version, deleted_at, created_at, updated_at)
ON work_checklist_items TO harpy_runtime;
GRANT UPDATE (item_text, completed, position, row_version, deleted_at, updated_at) ON work_checklist_items TO harpy_runtime;
GRANT INSERT (id, board_id, task_id, author_id, body, state, redaction_kind, redacted_by, redaction_reason, row_version, created_at, redacted_at)
ON work_comments TO harpy_runtime;
GRANT UPDATE (body, state, redaction_kind, redacted_by, redaction_reason, row_version, redacted_at)
ON work_comments TO harpy_runtime;
GRANT INSERT (id, owner_event_id, current_revision, lifecycle_state, row_version, created_at, updated_at)
ON work_recurrence_rules TO harpy_runtime;
GRANT UPDATE (current_revision, lifecycle_state, row_version, updated_at) ON work_recurrence_rules TO harpy_runtime;
GRANT INSERT (rule_id, revision, contract_version, dtstart, timezone_name, tzdb_version, rrule, exdates, effective_from, created_at)
ON work_recurrence_rule_revisions TO harpy_runtime;
GRANT INSERT (id, owner_id, audience, board_id, organization_id, title, description, starts_at, ends_at, start_date, end_date, timezone_name, rule_id, rule_revision, occurrence_start, tzdb_version, generated, generated_state, edited, live_reference_count, legal_hold, row_version, deleted_at, created_at, updated_at)
ON work_calendar_events TO harpy_runtime;
GRANT UPDATE (audience, board_id, organization_id, title, description, starts_at, ends_at, start_date, end_date, timezone_name, edited, row_version, deleted_at, updated_at)
ON work_calendar_events TO harpy_runtime;
GRANT INSERT (id, work_resource_kind, work_resource_id, target_owner, target_kind, target_id, target_version_id, usage, target_reference_epoch, permission_revision, state, row_version, created_at, released_at)
ON work_resource_links TO harpy_runtime;
GRANT INSERT (source_owner, aggregate_kind, aggregate_id, through_sequence, bootstrap_sequence, state, state_sha256, updated_at)
ON work_projection_streams TO harpy_runtime;
GRANT UPDATE (through_sequence, bootstrap_sequence, state, state_sha256, updated_at) ON work_projection_streams TO harpy_runtime;
GRANT INSERT (id, source_owner, source_kind, source_id, source_version_id, source_revision, permission_revision, event_sequence, summary, due_at, state, updated_at)
ON work_inbox_items TO harpy_runtime;
GRANT UPDATE (source_revision, permission_revision, event_sequence, summary, due_at, state, updated_at)
ON work_inbox_items TO harpy_runtime;
GRANT INSERT (id, source_owner, source_kind, source_id, source_version_id, source_revision, permission_revision, event_sequence, summary, starts_at, ends_at, start_date, end_date, state, updated_at)
ON work_calendar_items TO harpy_runtime;
GRANT UPDATE (source_revision, permission_revision, event_sequence, summary, starts_at, ends_at, start_date, end_date, state, updated_at)
ON work_calendar_items TO harpy_runtime;
GRANT INSERT (operation_id, operation_kind, actor_id, idempotency_key, request_sha256, aggregate_kind, state, aggregate_id, aggregate_version, response_status, response_document, occurred_at, completed_at)
ON work_operations TO harpy_runtime;
GRANT UPDATE (state, aggregate_id, aggregate_version, response_status, response_document, completed_at)
ON work_operations TO harpy_runtime;
GRANT INSERT (aggregate_kind, aggregate_id, event_sequence, aggregate_version, updated_at)
ON work_aggregate_sequences TO harpy_runtime;
GRANT UPDATE (event_sequence, aggregate_version, updated_at) ON work_aggregate_sequences TO harpy_runtime;
GRANT INSERT (audit_id, operation_id, actor_id, action, aggregate_kind, aggregate_id, aggregate_version, event_sequence, reason, occurred_at)
ON work_audit_entries TO harpy_runtime;

GRANT EXECUTE ON FUNCTION work_guard_comment_mutation_v1() TO harpy_runtime;
GRANT EXECUTE ON FUNCTION work_guard_recurrence_revision_v1() TO harpy_runtime;
GRANT EXECUTE ON FUNCTION work_guard_board_column_limit_v1() TO harpy_runtime;
GRANT EXECUTE ON FUNCTION work_guard_task_board_limit_v1() TO harpy_runtime;
GRANT EXECUTE ON FUNCTION work_guard_last_user_admin_v1() TO harpy_runtime;
