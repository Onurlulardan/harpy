package postgres

import _ "embed"

//go:embed retention/sql/tenant/lock_checkpoint.sql
var lockCheckpointSQL string

//go:embed retention/sql/tenant/find_receipt.sql
var findReceiptSQL string

//go:embed retention/sql/tenant/redact_content.sql
var redactContentSQL string

//go:embed retention/sql/tenant/purge_attempts.sql
var purgeAttemptsSQL string

//go:embed retention/sql/tenant/purge_receipts.sql
var purgeReceiptsSQL string

//go:embed retention/sql/tenant/purge_effects.sql
var purgeEffectsSQL string

//go:embed retention/sql/tenant/purge_hints.sql
var purgeHintsSQL string

//go:embed retention/sql/tenant/purge_deliveries.sql
var purgeDeliveriesSQL string

//go:embed retention/sql/tenant/purge_messages.sql
var purgeMessagesSQL string

//go:embed retention/sql/platform/purge_records.sql
var purgePlatformRecordsSQL string

//go:embed retention/sql/tenant/has_remaining.sql
var hasRemainingTenantSQL string

//go:embed retention/sql/platform/has_remaining.sql
var hasRemainingPlatformSQL string

//go:embed retention/sql/tenant/insert_receipt.sql
var insertReceiptSQL string

//go:embed retention/sql/tenant/update_checkpoint.sql
var updateCheckpointSQL string
