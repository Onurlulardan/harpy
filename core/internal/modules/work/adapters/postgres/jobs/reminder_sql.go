package jobs

import _ "embed"

//go:embed sql/tenant/load_due_reminders.sql
var loadDueRemindersSQL string

//go:embed sql/tenant/record_reminder.sql
var recordReminderSQL string
