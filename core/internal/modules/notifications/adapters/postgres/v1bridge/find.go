package v1bridge

import (
	"context"
	"crypto/sha256"
	"encoding/json"

	bridgeapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1bridge"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/google/uuid"
)

func (store *Store) FindIntent(ctx context.Context, producer string,
	effectHash [32]byte, recipient [16]byte,
) (bridgeapp.Intent, bool, error) {
	if !store.usable(ctx) || !validProducer(producer) || effectHash == [32]byte{} ||
		recipient == [16]byte{} {
		return bridgeapp.Intent{}, false, bridgeapp.ErrStorage
	}
	var value bridgeapp.Intent
	var effect, canonical, payloadDigest, payload []byte
	var recipientText, jobText string
	err := store.database.QueryRow(ctx, findIntentSQL, producer, effectHash[:],
		uuid.UUID(recipient)).Scan(&value.Producer, &effect, &recipientText,
		&value.ProducerGeneration, &value.NotificationGeneration, &canonical,
		&payloadDigest, &payload, &jobText, &value.State, &value.CreatedAt,
		&value.UpdatedAt, &value.Revision)
	if classifyNotFound(err) {
		return bridgeapp.Intent{}, false, nil
	}
	value.EffectKeyHash, err = digest(effect)
	if err == nil {
		value.CanonicalDigest, err = digest(canonical)
	}
	if err == nil {
		value.PayloadDigest, err = digest(payloadDigest)
	}
	if err == nil {
		value.RecipientID, err = domain.ParseID(recipientText)
	}
	if err == nil {
		value.JobID, err = domain.ParseID(jobText)
	}
	value.Payload = append([]byte(nil), payload...)
	if err != nil || !validStoredIntent(value, producer, effectHash,
		uuid.UUID(recipient)) {
		return bridgeapp.Intent{}, false, bridgeapp.ErrStorage
	}
	return value, true, nil
}

func validStoredIntent(value bridgeapp.Intent, producer string, effect [32]byte,
	recipient uuid.UUID,
) bool {
	validState := value.State == bridgeapp.StateQueued ||
		value.State == bridgeapp.StateAccepted || value.State == bridgeapp.StateFailed ||
		value.State == bridgeapp.StateOutcomeUnknown
	return value.Producer == producer && value.EffectKeyHash == effect &&
		uuid.UUID(value.RecipientID) == recipient && value.CanonicalDigest != [32]byte{} &&
		value.PayloadDigest == sha256.Sum256(value.Payload) && json.Valid(value.Payload) &&
		value.JobID.Valid() && value.ProducerGeneration > 0 &&
		value.NotificationGeneration > 0 && validState && value.Revision > 0 &&
		!value.CreatedAt.IsZero() && !value.UpdatedAt.Before(value.CreatedAt)
}

func digest(value []byte) ([32]byte, error) {
	var result [32]byte
	if len(value) != len(result) {
		return result, bridgeapp.ErrStorage
	}
	copy(result[:], value)
	return result, nil
}
