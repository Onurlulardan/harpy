package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type boundaryJobDocument struct {
	ID               uuid.UUID                    `json:"id"`
	EffectID         uuid.UUID                    `json:"effect_id"`
	SubscriptionID   uuid.UUID                    `json:"subscription_id"`
	Boundary         subscriptionapp.BoundaryKind `json:"boundary"`
	RunAtUnixMicro   int64                        `json:"run_at_unix_micro"`
	ExpectedRevision int64                        `json:"expected_revision"`
	PayloadRevision  int                          `json:"payload_revision"`
}

func (storage *store) listBoundaryJobs(
	ctx context.Context,
	authorityID, subscriptionID uuid.UUID,
) ([]subscriptionapp.JobIntent, error) {
	var document []byte
	if err := storage.queryRow(
		ctx, listBoundaryJobsSQL, authorityID, storage.tenantID, subscriptionID,
	).Scan(&document); err != nil {
		return nil, fmt.Errorf("list billing subscription boundary jobs: %w", err)
	}
	var records []boundaryJobDocument
	if err := json.Unmarshal(document, &records); err != nil {
		return nil, fmt.Errorf("decode billing subscription boundary jobs: %w", err)
	}
	result := make([]subscriptionapp.JobIntent, 0, len(records))
	for _, record := range records {
		result = append(result, subscriptionapp.JobIntent{
			ID: record.ID, EffectID: record.EffectID,
			SubscriptionID: record.SubscriptionID, Boundary: record.Boundary,
			RunAt:            time.UnixMicro(record.RunAtUnixMicro).UTC(),
			ExpectedRevision: record.ExpectedRevision,
			PayloadRevision:  record.PayloadRevision,
		})
	}
	return result, nil
}
