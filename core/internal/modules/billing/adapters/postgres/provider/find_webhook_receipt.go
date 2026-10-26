package provider

import (
	"context"
	"fmt"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func (storage *ingressRepository) findReceipt(
	ctx context.Context,
	target providerTarget,
	acceptance providerapp.WebhookAcceptance,
) (providerapp.WebhookAcceptDecision, bool, error) {
	var receiptID, tenantID, mappingID uuid.UUID
	var bodySHA []byte
	var eventID string
	err := storage.database.queryRow(
		ctx, findWebhookReceiptSQL, target.providerAccountID,
		acceptance.Verified().EventID(),
	).Scan(&receiptID, &tenantID, &mappingID, &bodySHA, &eventID)
	if storage.database.notFound(err) {
		return providerapp.WebhookAcceptDecision{}, false, nil
	}
	if err != nil {
		return providerapp.WebhookAcceptDecision{}, false,
			fmt.Errorf("find billing webhook receipt: %w", err)
	}
	wantedSHA := acceptance.Body().BodySHA256()
	if receiptID == uuid.Nil || tenantID != target.tenantID ||
		mappingID != target.mappingID || eventID != acceptance.Verified().EventID() ||
		len(bodySHA) != len(wantedSHA) || string(bodySHA) != string(wantedSHA[:]) {
		return providerapp.WebhookAcceptDecision{}, false, providerapp.ErrWebhookConflict
	}
	return providerapp.WebhookAcceptDecision{
		ReceiptID: receiptID, TenantID: tenantID, Created: false,
	}, true, nil
}
