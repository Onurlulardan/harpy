package providerapp

type WebhookProcessState string

const (
	WebhookApplied         WebhookProcessState = "applied"
	WebhookDuplicate       WebhookProcessState = "duplicate"
	WebhookReconcileQueued WebhookProcessState = "reconcile_queued"
)

type WebhookProcessDecision struct {
	State WebhookProcessState
}

func (decision WebhookProcessDecision) valid() bool {
	switch decision.State {
	case WebhookApplied, WebhookDuplicate, WebhookReconcileQueued:
		return true
	default:
		return false
	}
}

func (decision WebhookProcessDecision) validFor(ordering EventOrdering) bool {
	if !decision.valid() || !ordering.valid() {
		return false
	}
	return ordering.Mode() != OrderingRequiresReconciliation ||
		decision.State == WebhookDuplicate || decision.State == WebhookReconcileQueued
}
