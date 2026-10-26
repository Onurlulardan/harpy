package v1adoption

import _ "embed"

//go:embed sql/tenant/adopt_board.sql
var adoptBoardSQL string

//go:embed sql/tenant/adopt_column.sql
var adoptColumnSQL string

//go:embed sql/tenant/adopt_membership.sql
var adoptMembershipSQL string

//go:embed sql/tenant/adopt_task.sql
var adoptTaskSQL string

//go:embed sql/tenant/adopt_task_compatibility_metadata.sql
var adoptTaskCompatibilityMetadataSQL string

//go:embed sql/tenant/adopt_assignee.sql
var adoptAssigneeSQL string

//go:embed sql/tenant/adopt_label.sql
var adoptLabelSQL string

//go:embed sql/tenant/adopt_task_label.sql
var adoptTaskLabelSQL string

//go:embed sql/tenant/adopt_checklist.sql
var adoptChecklistSQL string

//go:embed sql/tenant/adopt_comment.sql
var adoptCommentSQL string

//go:embed sql/tenant/adopt_calendar.sql
var adoptCalendarSQL string

//go:embed sql/tenant/adopt_recurrence_rule.sql
var adoptRecurrenceRuleSQL string

//go:embed sql/tenant/adopt_recurrence_revision.sql
var adoptRecurrenceRevisionSQL string

//go:embed sql/tenant/adopt_form_link.sql
var adoptFormLinkSQL string

//go:embed sql/tenant/adopt_workflow_projection.sql
var adoptProjectionSQL string
