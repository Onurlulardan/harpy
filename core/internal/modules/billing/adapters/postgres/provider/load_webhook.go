package provider

import (
	"context"
	"fmt"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
	"github.com/google/uuid"
)

func (storage *processorStore) LoadWebhook(
	ctx context.Context,
	receiptID uuid.UUID,
) (providerapp.StoredWebhook, error) {
	if !storage.usable(ctx) || receiptID == uuid.Nil {
		return providerapp.StoredWebhook{}, providerapp.ErrWebhookProcessing
	}
	row, err := storage.loadWebhookRow(ctx, receiptID)
	if err != nil {
		return providerapp.StoredWebhook{}, err
	}
	if !row.state.valid {
		return providerapp.StoredWebhook{}, providerapp.ErrWebhookProcessing
	}
	if row.state.value == "received" {
		var marked uuid.UUID
		err = storage.database.queryRow(
			ctx, markWebhookProcessingSQL, row.authorityID, storage.tenantID, receiptID,
		).Scan(&marked)
		if err != nil || marked != receiptID {
			return providerapp.StoredWebhook{}, fmt.Errorf(
				"mark billing webhook processing: %w", err,
			)
		}
	} else if row.state.value != "processing" && row.state.value != "applied" &&
		row.state.value != "reconcile_required" {
		return providerapp.StoredWebhook{}, providerapp.ErrWebhookProcessing
	}
	return row.stored(receiptID, storage.tenantID)
}

type webhookRow struct {
	authorityID, mappingID, providerAccountID             uuid.UUID
	subscriptionID, invoiceID                             *uuid.UUID
	providerKey, accountID, objectKind, objectID, eventID string
	authentication, nonce, encrypted, wrapped, bodySHA    []byte
	verification                                          int
	algorithm, keyID, state                               nullableString
	occurredAt, signatureAt, receivedAt                   nullableTime
}

func (storage *processorStore) loadWebhookRow(
	ctx context.Context,
	receiptID uuid.UUID,
) (webhookRow, error) {
	var row webhookRow
	err := storage.database.queryRow(
		ctx, loadWebhookSQL, storage.tenantID, receiptID,
	).Scan(
		&row.authorityID, &row.mappingID, &row.providerAccountID,
		&row.providerKey, &row.accountID, &row.objectKind,
		&row.objectID, &row.subscriptionID, &row.invoiceID, &row.eventID,
		&row.authentication, &row.verification, &row.occurredAt, &row.signatureAt,
		&row.algorithm, &row.nonce, &row.encrypted, &row.wrapped, &row.keyID,
		&row.bodySHA, &row.receivedAt, &row.state,
	)
	if err != nil {
		return webhookRow{}, fmt.Errorf("load billing webhook receipt: %w", err)
	}
	return row, nil
}

func (row webhookRow) stored(
	receiptID uuid.UUID,
	tenantID uuid.UUID,
) (providerapp.StoredWebhook, error) {
	key, err := domainprovider.NewKey(row.providerKey)
	if err != nil || !row.algorithm.valid || !row.keyID.valid || !row.state.valid ||
		!row.occurredAt.valid || !row.signatureAt.valid || !row.receivedAt.valid ||
		len(row.authentication) != 32 || len(row.bodySHA) != 32 {
		return providerapp.StoredWebhook{}, providerapp.ErrWebhookProcessing
	}
	var authentication, bodySHA [32]byte
	copy(authentication[:], row.authentication)
	copy(bodySHA[:], row.bodySHA)
	verified, err := providerapp.NewVerifiedWebhook(
		row.eventID, row.accountID, row.objectKind, row.objectID,
		row.occurredAt.value, row.signatureAt.value, row.verification, authentication,
	)
	if err != nil {
		return providerapp.StoredWebhook{}, err
	}
	body, err := providerapp.NewProtectedBody(
		row.algorithm.value, row.nonce, row.encrypted,
		row.keyID.value, row.wrapped, bodySHA,
	)
	if err != nil {
		return providerapp.StoredWebhook{}, err
	}
	return providerapp.NewStoredWebhook(
		receiptID, tenantID, key, verified, body, row.receivedAt.value,
	)
}
