package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

const (
	EmailDeliveryJobKind    = "notifications.email.delivery"
	TelegramDeliveryJobKind = "notifications.telegram.delivery"
)

func (store *Store) EnqueueDelivery(ctx context.Context,
	job ports.DeliveryJob,
) error {
	if !store.usable(ctx, true) || store.enqueue == nil || store.tenantID == [16]byte{} ||
		!job.ID.Valid() || !job.DeliveryID.Valid() || job.EffectKey == "" ||
		job.RunAt.IsZero() {
		return ErrInvalid
	}
	kind := EmailDeliveryJobKind
	if job.Channel == domain.ChannelTelegram {
		kind = TelegramDeliveryJobKind
	} else if job.Channel != domain.ChannelEmail {
		return ErrInvalid
	}
	tenant := store.tenantID.String()
	reference := "notification-delivery://tenant/" + tenant +
		"/delivery/" + job.DeliveryID.String()
	digest := sha256.Sum256([]byte("notifications.delivery.v1\x00" + reference))
	authority := store.authority
	value, err := jobsv1.NewEnqueue(jobsv1.Enqueue{JobID: job.ID.String(),
		Kind: kind, PayloadRevision: 1, TenantID: &tenant,
		Authority: jobsv1.Authority{ScopeKind: jobsv1.ScopeTenant, ScopeID: tenant,
			Domain: authority.domain, CutoverGeneration: authority.generation,
			WriterOrigin: jobsv1.WriterV2, GroupID: authority.groupID,
			GroupGeneration: authority.groupGen}, PayloadReference: reference,
		PayloadSHA256: hex.EncodeToString(digest[:]), RunAt: job.RunAt.UTC().Format(time.RFC3339Nano),
		Priority: 100, DedupeKey: job.DeliveryID.String(),
		EffectIdempotencyKey: &job.EffectKey, MaximumAttempts: 8}, store.jobs)
	if err != nil {
		return err
	}
	storedID, err := store.enqueue(ctx, value)
	if err != nil || storedID != job.ID.String() {
		return ErrConflict
	}
	return nil
}
