package httpapi

type WebhookRequest struct {
	Provider string
	Headers  map[string][]string
	Body     []byte
}

type WebhookAcceptance struct {
	Duplicate bool
	ReceiptID string
}
