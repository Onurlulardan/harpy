package provider

import (
	"context"
	"fmt"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func (storage *ingressRepository) authorityID(
	ctx context.Context,
	tenantID uuid.UUID,
) (uuid.UUID, error) {
	var authorityID uuid.UUID
	err := storage.database.queryRow(
		ctx, resolveTenantAuthoritySQL, tenantID,
	).Scan(&authorityID)
	if err != nil || authorityID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("resolve billing webhook authority: %w", err)
	}
	return authorityID, nil
}

func (storage *ingressRepository) insertReceipt(
	ctx context.Context,
	target providerTarget,
	authorityID uuid.UUID,
	acceptance providerapp.WebhookAcceptance,
) (bool, error) {
	verified, body := acceptance.Verified(), acceptance.Body()
	authentication := verified.AuthenticationSHA256()
	bodySHA := body.BodySHA256()
	receivedAt := acceptance.ReceivedAt()
	rows, err := storage.database.execute(
		ctx, insertWebhookReceiptSQL,
		acceptance.ReceiptID(), authorityID, target.tenantID, target.providerAccountID,
		target.mappingID, verified.EventID(), bodySHA[:], authentication[:],
		verified.VerificationRevision(), verified.OccurredAt(),
		verified.SignatureIssuedAt(), body.Algorithm(), body.Nonce(), body.Ciphertext(),
		body.WrappedDEK(), body.KeyID(), receivedAt, receivedAt,
		receivedAt.Add(rawBodyRetention), receivedAt.Add(receiptRetention),
	)
	if err != nil {
		return false, fmt.Errorf("insert billing webhook receipt: %w", err)
	}
	return rows == 1, nil
}
