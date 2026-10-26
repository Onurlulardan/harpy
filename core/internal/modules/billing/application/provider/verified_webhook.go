package providerapp

import "time"

type VerifiedWebhook struct {
	eventID           string
	accountID         string
	objectKind        string
	objectID          string
	occurredAt        time.Time
	signatureIssuedAt time.Time
	verification      int
	authenticationSHA [32]byte
}

func NewVerifiedWebhook(
	eventID string,
	accountID string,
	objectKind string,
	objectID string,
	occurredAt time.Time,
	signatureIssuedAt time.Time,
	verificationRevision int,
	authenticationSHA256 [32]byte,
) (VerifiedWebhook, error) {
	canonicalOccurred, occurredErr := canonicalInstant(occurredAt)
	canonicalSignature, signatureErr := canonicalInstant(signatureIssuedAt)
	if !validExternalIdentifier(eventID) || !validExternalIdentifier(accountID) ||
		!validExternalIdentifier(objectKind) || !validExternalIdentifier(objectID) ||
		occurredErr != nil || signatureErr != nil || verificationRevision < 1 {
		return VerifiedWebhook{}, ErrWebhookMalformed
	}
	if authenticationSHA256 == [32]byte{} {
		return VerifiedWebhook{}, ErrWebhookMalformed
	}
	return VerifiedWebhook{
		eventID: eventID, accountID: accountID, objectKind: objectKind,
		objectID: objectID, occurredAt: canonicalOccurred,
		signatureIssuedAt: canonicalSignature, verification: verificationRevision,
		authenticationSHA: authenticationSHA256,
	}, nil
}

func (webhook VerifiedWebhook) EventID() string       { return webhook.eventID }
func (webhook VerifiedWebhook) AccountID() string     { return webhook.accountID }
func (webhook VerifiedWebhook) ObjectKind() string    { return webhook.objectKind }
func (webhook VerifiedWebhook) ObjectID() string      { return webhook.objectID }
func (webhook VerifiedWebhook) OccurredAt() time.Time { return webhook.occurredAt }
func (webhook VerifiedWebhook) SignatureIssuedAt() time.Time {
	return webhook.signatureIssuedAt
}
func (webhook VerifiedWebhook) VerificationRevision() int { return webhook.verification }
func (webhook VerifiedWebhook) AuthenticationSHA256() [32]byte {
	return webhook.authenticationSHA
}

func (webhook VerifiedWebhook) valid() bool {
	validated, err := NewVerifiedWebhook(
		webhook.eventID, webhook.accountID, webhook.objectKind, webhook.objectID,
		webhook.occurredAt, webhook.signatureIssuedAt, webhook.verification,
		webhook.authenticationSHA,
	)
	return err == nil && validated == webhook
}
