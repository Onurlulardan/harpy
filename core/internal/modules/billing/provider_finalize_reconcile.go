package billing

import (
	"time"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func providerReconciliationChild(
	result providerapp.Finalization,
) (platformasync.DatabaseJobChildInput, error) {
	job, err := providerapp.NewProviderCommandReconciliationJob(
		result.Provider, result.CommandID, result.RequestSHA256,
		result.OccurredAt.Add(time.Second),
	)
	if err != nil {
		return platformasync.DatabaseJobChildInput{}, err
	}
	return platformasync.DatabaseJobChildInput{
		JobID: result.ActiveJobID, Kind: job.Policy().Kind,
		PayloadRevision: job.Policy().PayloadRevision, RunAt: job.RunAt(),
		PayloadReference: job.PayloadReference(), PayloadSHA256: job.PayloadSHA256(),
		DedupeKey: job.DedupeKey(), EffectIdempotencyKey: job.EffectKey(),
		Priority: 100,
	}, nil
}
