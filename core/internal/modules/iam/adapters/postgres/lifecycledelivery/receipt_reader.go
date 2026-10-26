package lifecycledelivery

import (
	"context"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
)

func (store *store) readReceipt(ctx context.Context,
	command lifecycledeliveryapp.ClaimCommand,
) (lifecycledeliveryapp.StoredResult, error) {
	var result lifecycledeliveryapp.StoredResult
	var digest []byte
	err := store.database.QueryRow(ctx, readReceiptSQL, command.ArtifactRef).Scan(
		&result.ReceiptID, &result.ArtifactRef, &result.EffectKey,
		&result.MessageID, &result.Outcome, &result.Code,
		&result.ProviderReference, &digest, &result.CompletedAt,
	)
	defer clear(digest)
	if err != nil {
		return lifecycledeliveryapp.StoredResult{}, err
	}
	if len(digest) != 32 {
		return lifecycledeliveryapp.StoredResult{}, lifecycledeliveryapp.ErrInvalid
	}
	copy(result.ResultSHA256[:], digest)
	return result, nil
}
