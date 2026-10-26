package invitationdelivery

import (
	"context"

	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
	"github.com/google/uuid"
)

func (store *store) readReceipt(
	ctx context.Context,
	command invitationdeliveryapp.ClaimCommand,
) (invitationdeliveryapp.StoredResult, error) {
	return store.scanReceipt(
		ctx, readReceiptSQL, command.ArtifactRef,
	)
}

func (store *store) scanReceipt(
	ctx context.Context,
	statement string,
	arguments ...any,
) (invitationdeliveryapp.StoredResult, error) {
	var result invitationdeliveryapp.StoredResult
	var digest []byte
	err := store.database.QueryRow(ctx, statement, arguments...).Scan(
		&result.ReceiptID, &result.ArtifactRef, &result.EffectKey,
		&result.MessageID, &result.Outcome, &result.Code,
		&result.ProviderReference, &digest, &result.CompletedAt,
	)
	if err != nil {
		return invitationdeliveryapp.StoredResult{}, err
	}
	if len(digest) != 32 {
		return invitationdeliveryapp.StoredResult{}, invitationdeliveryapp.ErrInvalid
	}
	copy(result.ResultSHA256[:], digest)
	return result, nil
}

func artifactID(reference string) (uuid.UUID, error) {
	parsed, err := deliveryReference(reference)
	if err != nil {
		return uuid.Nil, err
	}
	return parsed.ArtifactID(), nil
}
