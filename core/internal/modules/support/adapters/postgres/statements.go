package postgres

import _ "embed"

var (
	//go:embed sql/platform/submit_contact.sql
	submitContactSQL string
	//go:embed sql/platform/list_contacts.sql
	listContactsSQL string
	//go:embed sql/platform/contact_stats.sql
	contactStatsSQL string
	//go:embed sql/platform/get_contact.sql
	getContactSQL string
	//go:embed sql/platform/update_contact.sql
	updateContactSQL string
	//go:embed sql/platform/create_bug.sql
	createBugSQL string
	//go:embed sql/platform/list_bugs.sql
	listBugsSQL string
	//go:embed sql/platform/bug_stats.sql
	bugStatsSQL string
	//go:embed sql/platform/get_bug.sql
	getBugSQL string
	//go:embed sql/platform/update_bug.sql
	updateBugSQL string
)
