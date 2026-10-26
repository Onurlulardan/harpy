package subscriptionapp

const (
	OperationCancelAtPeriodEnd  = "billing.subscription.cancel_at_period_end"
	OperationChangeAtPeriodEnd  = "billing.subscription.change_at_period_end"
	OperationClaimTrial         = "billing.subscription.claim_trial"
	OperationExecuteDue         = "billing.subscription.execute_due_boundary"
	OperationPlatformTransition = "billing.subscription.platform_transition"
	OperationResumeCancellation = "billing.subscription.resume_cancellation"
	OperationUpdateFromProvider = "billing.subscription.update_from_provider"
)

const (
	permissionBillingManage   = "billing.manage"
	permissionPlatformManage  = "platform.billing.manage"
	outboxSubscriptionChanged = "billing.subscription.changed"
	outboxEntitlementsChanged = "billing.entitlements.changed"
)
