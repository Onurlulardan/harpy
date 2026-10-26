package application

import (
	"context"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *RecoveryProtectionService) AcknowledgeKit(
	ctx context.Context,
	operationID, actorUserID, sessionID, artifactSHA256 string,
	ackProof []byte,
) error {
	digest, err := hex.DecodeString(artifactSHA256)
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(operationID) || !validRecoveryUUID(actorUserID) ||
		!validRecoveryUUID(sessionID) || err != nil || len(digest) != 32 ||
		len(ackProof) == 0 {
		return api.ErrInvalid
	}
	err = service.store.AcknowledgeRecoveryKit(
		ctx, operationID, actorUserID, sessionID, digest,
		append([]byte(nil), ackProof...), service.clock.Now())
	if err == nil {
		return nil
	}
	return mapRecoveryStoreError(err)
}
