package postgres

import _ "embed"

var (
	//go:embed sql/tenant/lock_smtp_settings.sql
	lockSMTPSettingsSQL string
	//go:embed sql/tenant/latest_smtp_settings_revision.sql
	latestSMTPSettingsRevisionSQL string
	//go:embed sql/tenant/find_smtp_settings_receipt.sql
	findSMTPSettingsReceiptSQL string
	//go:embed sql/tenant/insert_smtp_settings_receipt.sql
	insertSMTPSettingsReceiptSQL string
	//go:embed sql/tenant/find_smtp_test_intent.sql
	findSMTPTestIntentSQL string
	//go:embed sql/tenant/insert_smtp_test_intent.sql
	insertSMTPTestIntentSQL string
)
