package provider

import (
	"context"
	"fmt"
	"time"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

const rawBodyRetention = 30 * 24 * time.Hour
const receiptRetention = 400 * 24 * time.Hour

func (storage *ingressRepository) accept(
	ctx context.Context,
	target providerTarget,
	acceptance providerapp.WebhookAcceptance,
	jobs JobWriter,
) (providerapp.WebhookAcceptDecision, error) {
	if storage == nil || !storage.database.usable() || jobs == nil ||
		!validAcceptance(target, acceptance) {
		return providerapp.WebhookAcceptDecision{}, providerapp.ErrWebhookDecision
	}
	existing, found, err := storage.findReceipt(ctx, target, acceptance)
	if err != nil || found {
		return existing, err
	}
	authorityID, err := storage.authorityID(ctx, target.tenantID)
	if err != nil {
		return providerapp.WebhookAcceptDecision{}, err
	}
	created, err := storage.insertReceipt(ctx, target, authorityID, acceptance)
	if err != nil {
		return providerapp.WebhookAcceptDecision{}, err
	}
	if !created {
		existing, found, err = storage.findReceipt(ctx, target, acceptance)
		if err != nil || !found {
			return providerapp.WebhookAcceptDecision{}, fmt.Errorf(
				"read concurrent billing webhook receipt: %w", err,
			)
		}
		return existing, nil
	}
	if err := jobs.EnqueueProviderJob(
		ctx, acceptance.ReceiptID(), acceptance.ProcessorJob(),
	); err != nil {
		return providerapp.WebhookAcceptDecision{}, err
	}
	return providerapp.WebhookAcceptDecision{
		ReceiptID: acceptance.ReceiptID(), TenantID: target.tenantID, Created: true,
	}, nil
}

func validAcceptance(
	target providerTarget,
	acceptance providerapp.WebhookAcceptance,
) bool {
	verified, body := acceptance.Verified(), acceptance.Body()
	return acceptance.ReceiptID() != uuid.Nil &&
		acceptance.ReceiptID().Version() == uuid.Version(7) &&
		acceptance.Provider().Valid() && target.tenantID != uuid.Nil &&
		target.mappingID != uuid.Nil && target.providerAccountID != uuid.Nil &&
		verified.EventID() != "" &&
		verified.AccountID() != "" && verified.ObjectKind() != "" &&
		verified.ObjectID() != "" && !verified.OccurredAt().After(acceptance.ReceivedAt()) &&
		!verified.SignatureIssuedAt().After(acceptance.ReceivedAt()) &&
		body.Algorithm() == "AES-256-GCM" && len(body.Nonce()) == 12 &&
		len(body.Ciphertext()) >= 17 && body.BodySHA256() != [32]byte{}
}
