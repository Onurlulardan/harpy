package providerapp

import "github.com/google/uuid"

type WebhookAcceptDecision struct {
	ReceiptID uuid.UUID
	TenantID  uuid.UUID
	Created   bool
}

func (decision WebhookAcceptDecision) valid(requestedReceiptID uuid.UUID) bool {
	if decision.ReceiptID == uuid.Nil || decision.TenantID == uuid.Nil {
		return false
	}
	return !decision.Created || decision.ReceiptID == requestedReceiptID
}

type ReceiveResult struct {
	ReceiptID uuid.UUID
	Created   bool
}
