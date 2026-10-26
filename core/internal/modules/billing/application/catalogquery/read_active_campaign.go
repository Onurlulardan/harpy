package catalogquery

import (
	"context"
	"fmt"
	"time"
)

func NewActiveCampaignRequest(instant time.Time) (ActiveCampaignRequest, error) {
	if !validInstant(instant) {
		return ActiveCampaignRequest{}, ErrRequest
	}
	return ActiveCampaignRequest{instant: instant}, nil
}

func (queries *Queries) ReadActiveCampaign(
	ctx context.Context,
	request ActiveCampaignRequest,
) (*CampaignSnapshot, error) {
	if !queries.usable(ctx) || !validInstant(request.instant) {
		return nil, ErrRequest
	}
	record, found, err := queries.store.ReadActiveCampaign(
		ctx, CampaignStoreQuery{Instant: request.instant},
	)
	if err != nil {
		return nil, fmt.Errorf("read active billing campaign: %w", err)
	}
	if !found {
		return nil, nil
	}
	if !activeAt(record.StartsAt, record.EndsAt, request.instant) {
		return nil, ErrResult
	}
	result, valid := projectCampaign(record, true)
	if !valid {
		return nil, ErrResult
	}
	return &result, nil
}
