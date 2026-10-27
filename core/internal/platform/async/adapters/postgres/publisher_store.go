package postgres

import (
	"context"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

const (
	maximumOutboxAuthorityPageSize = 256
	maximumPublishBatchSize        = 50
	maximumPublishRetryDelay       = 15 * time.Minute
	publishLeaseDuration           = 60 * time.Second
)

type PublisherStore interface {
	AssertOutboxCatalog(context.Context, eventsv1.ProducerCatalog) error
	BeginOutboxAuthorityPage(int) (OutboxAuthorityPageToken, error)
	ListOutboxAuthorities(context.Context, OutboxAuthorityPageToken) (OutboxAuthorityPage, error)
	ClaimOutbox(
		context.Context, OutboxAuthority, uuid.UUID, int, eventsv1.ProducerCatalog,
	) ([]PublishClaim, error)
	HeartbeatPublish(context.Context, PublishClaim) (PublishLease, error)
	ConfirmPublish(context.Context, PublishClaim, string) (time.Time, error)
	ReleasePublish(
		context.Context, PublishClaim, time.Duration, string, string,
	) (time.Time, error)
	ParkStaleOutbox(context.Context, OutboxAuthority, int) (int64, error)
	RecoverExpiredOutbox(context.Context, OutboxAuthority, int) (int64, error)
}

var _ PublisherStore = (*workerRepository)(nil)
