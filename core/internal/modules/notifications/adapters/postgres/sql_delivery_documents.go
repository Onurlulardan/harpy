package postgres

import _ "embed"

var (
	//go:embed sql/tenant/insert_delivery.sql
	insertDeliverySQL string
	//go:embed sql/tenant/get_delivery.sql
	getDeliverySQL string
	//go:embed sql/tenant/claim_delivery.sql
	claimDeliverySQL string
	//go:embed sql/tenant/update_delivery.sql
	updateDeliverySQL string
	//go:embed sql/tenant/insert_delivery_attempt.sql
	insertDeliveryAttemptSQL string
	//go:embed sql/tenant/insert_delivery_receipt.sql
	insertDeliveryReceiptSQL string
	//go:embed sql/tenant/claim_delivery_work.sql
	claimDeliveryWorkSQL string
	//go:embed sql/tenant/get_worker_message.sql
	getWorkerMessageSQL string
	//go:embed sql/tenant/resolve_notification_recipient.sql
	resolveNotificationRecipientSQL string
	//go:embed sql/tenant/get_provider_version.sql
	getProviderVersionSQL string
	//go:embed sql/tenant/reconcile_stale_delivery.sql
	reconcileStaleDeliverySQL string
	//go:embed sql/tenant/record_provider_receipt.sql
	recordProviderReceiptSQL string
)
