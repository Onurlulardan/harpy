package v1bridge

import (
	"context"
	"encoding/json"

	bridgeapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1bridge"
)

func (store *Store) InsertIntent(ctx context.Context, value bridgeapp.Intent) error {
	if !store.usable(ctx) || !validIntent(value) {
		return bridgeapp.ErrStorage
	}
	result, err := store.database.Execute(ctx, insertIntentSQL, value.Producer,
		value.EffectKeyHash[:], value.RecipientID.String(), value.ProducerGeneration,
		value.NotificationGeneration, value.CanonicalDigest[:], value.PayloadDigest[:],
		value.Payload, value.JobID.String(), string(value.State), value.CreatedAt.UTC(),
		value.UpdatedAt.UTC(), value.Revision)
	if err != nil || result.RowsAffected() != 1 {
		return bridgeapp.ErrStorage
	}
	return nil
}

func (store *Store) Enqueue(ctx context.Context, job bridgeapp.Job) error {
	if !store.usable(ctx) || job.TenantID.String() != store.tenantID.String() {
		return bridgeapp.ErrStorage
	}
	return store.enqueuer.EnqueueNotificationV1Bridge(ctx, job)
}

func validIntent(value bridgeapp.Intent) bool {
	return validProducer(value.Producer) && value.EffectKeyHash != [32]byte{} &&
		value.CanonicalDigest != [32]byte{} && value.PayloadDigest != [32]byte{} &&
		value.RecipientID.Valid() && value.JobID.IsV7() &&
		value.ProducerGeneration > 0 && value.NotificationGeneration > 0 &&
		value.State == bridgeapp.StateQueued && value.Revision == 1 &&
		!value.CreatedAt.IsZero() && value.UpdatedAt.Equal(value.CreatedAt) &&
		len(value.Payload) <= 64*1024 && json.Valid(value.Payload)
}
