package postgres

import _ "embed"

var (
	//go:embed sql/tenant/lock_effect.sql
	lockEffectSQL string
	//go:embed sql/tenant/find_effect.sql
	findEffectSQL string
	//go:embed sql/tenant/insert_effect.sql
	insertEffectSQL string
	//go:embed sql/tenant/update_effect_outcome.sql
	updateEffectOutcomeSQL string
	//go:embed sql/tenant/next_inbox_revision.sql
	nextInboxRevisionSQL string
	//go:embed sql/tenant/insert_message.sql
	insertMessageSQL string
	//go:embed sql/tenant/get_message.sql
	getMessageSQL string
	//go:embed sql/tenant/list_messages.sql
	listMessagesSQL string
	//go:embed sql/tenant/count_unread.sql
	countUnreadSQL string
	//go:embed sql/tenant/update_message.sql
	updateMessageSQL string
	//go:embed sql/tenant/mark_all_read.sql
	markAllReadSQL string
	//go:embed sql/tenant/current_inbox_revision.sql
	currentInboxRevisionSQL string
	//go:embed sql/tenant/get_preference.sql
	getPreferenceSQL string
	//go:embed sql/tenant/list_preferences.sql
	listPreferencesSQL string
	//go:embed sql/tenant/put_preference.sql
	putPreferenceSQL string
	//go:embed sql/tenant/get_published_template.sql
	getPublishedTemplateSQL string
	//go:embed sql/tenant/insert_template.sql
	insertTemplateSQL string
	//go:embed sql/tenant/insert_template_version.sql
	insertTemplateVersionSQL string
	//go:embed sql/tenant/publish_template.sql
	publishTemplateSQL string
)
