package subscription

import (
	"context"
	"unicode/utf8"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

type BoundaryArtifact struct {
	PayloadReference string
	PayloadSHA256    [32]byte
}

func (artifact BoundaryArtifact) valid() bool {
	return utf8.ValidString(artifact.PayloadReference) &&
		len(artifact.PayloadReference) >= 1 && len(artifact.PayloadReference) <= 2048 &&
		artifact.PayloadSHA256 != [32]byte{}
}

type Effects interface {
	EnqueueBoundary(
		context.Context,
		subscriptionapp.JobIntent,
	) (BoundaryArtifact, error)
	CancelBoundary(context.Context, subscriptionapp.JobCancellation) error
	PublishSubscription(
		context.Context,
		subscriptionapp.OutboxIntent,
		subscriptionapp.SubscriptionSnapshot,
	) error
}
