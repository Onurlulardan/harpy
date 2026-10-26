package provider

import (
	"context"
	"fmt"
	"time"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func (storage *processorStore) queueReconciliation(
	ctx context.Context,
	row webhookRow,
	application providerapp.WebhookApplication,
	processedAt time.Time,
) (providerapp.WebhookProcessDecision, error) {
	jobID, err := storage.nextID()
	if err != nil || jobID == uuid.Nil || jobID.Version() != uuid.Version(7) {
		return providerapp.WebhookProcessDecision{}, providerapp.ErrProviderDependency
	}
	if err := storage.jobs.EnqueueProviderJob(
		ctx, jobID, application.ReconciliationJob(),
	); err != nil {
		return providerapp.WebhookProcessDecision{}, err
	}
	var receiptID uuid.UUID
	err = storage.database.queryRow(
		ctx, markWebhookReconcileSQL, row.authorityID, storage.tenantID,
		application.ReceiptID(), processedAt,
	).Scan(&receiptID)
	if err != nil || receiptID != application.ReceiptID() {
		return providerapp.WebhookProcessDecision{}, fmt.Errorf(
			"mark billing webhook reconciliation: %w", err,
		)
	}
	return providerapp.WebhookProcessDecision{State: providerapp.WebhookReconcileQueued}, nil
}
