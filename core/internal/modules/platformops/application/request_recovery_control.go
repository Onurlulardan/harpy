package application

import (
	"context"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *RecoveryProtectionService) RequestControl(
	ctx context.Context,
	request api.RecoveryControlRequest,
) (api.RecoveryControlOperation, error) {
	digest, err := hex.DecodeString(request.RequestSHA256)
	claimRequired := request.Kind == "enable" || request.Kind == "rotate"
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(request.OperationID) ||
		!validRecoveryUUID(request.EnvironmentID) ||
		!validRecoveryUUID(request.ActorUserID) || len(digest) != 32 ||
		(request.Kind != "enable" && request.Kind != "rotate" && request.Kind != "disable") ||
		(claimRequired && (request.ClaimAlgorithm != api.RecoveryClaimAlgorithm ||
			len(request.ClaimPublicKey) != 32)) ||
		(!claimRequired && (request.ClaimAlgorithm != "" || len(request.ClaimPublicKey) != 0)) {
		return api.RecoveryControlOperation{}, api.ErrInvalid
	}
	record, err := service.store.RequestRecoveryControl(ctx,
		ports.RecoveryControlRequestRecord{
			OperationID: request.OperationID, EnvironmentID: request.EnvironmentID,
			ActorUserID: request.ActorUserID, IdempotencyKey: request.IdempotencyKey,
			RequestSHA256: digest, Kind: request.Kind,
			ClaimAlgorithm: request.ClaimAlgorithm,
			ClaimPublicKey: append([]byte(nil), request.ClaimPublicKey...),
			OccurredAt:     service.clock.Now(),
		})
	if err != nil {
		return api.RecoveryControlOperation{}, mapRecoveryStoreError(err)
	}
	return recoveryControlView(record), nil
}
