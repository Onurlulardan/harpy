package postrestore

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/google/uuid"
)

func (service *Service) Challenge(
	ctx context.Context, environmentID, restoreOperationID uuid.UUID, origin string,
) (ChallengeResult, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || environmentID == uuid.Nil ||
		restoreOperationID == uuid.Nil ||
		!service.validOrigin(origin) {
		return ChallengeResult{}, ErrDenied
	}
	binding, err := service.store.Binding(
		ctx, environmentID, restoreOperationID,
	)
	if err != nil {
		return ChallengeResult{}, mapStoreError(err)
	}
	operationID, err := service.newUUID()
	if err != nil || operationID == uuid.Nil {
		return ChallengeResult{}, ErrUnavailable
	}
	issued, err := service.security.Challenges.Generate()
	if err != nil {
		return ChallengeResult{}, ErrUnavailable
	}
	defer issued.Destroy()
	now := service.clock().UTC()
	expiresAt := now.Add(5 * time.Minute)
	var record ChallengeRecord
	err = issued.RevealOnce(func(wire []byte) error {
		challenge := string(wire)
		digest := challengeContextDigest(environmentID, restoreOperationID,
			binding, service.rpID, origin, challenge)
		record, err = service.store.PersistChallenge(ctx, ChallengeCommand{
			OperationID: operationID, EnvironmentID: environmentID,
			RestoreOperationID: restoreOperationID, Challenge: challenge,
			RPID: service.rpID, Origin: origin, ContextSHA256: digest,
			ExpiresAt: expiresAt, OccurredAt: now,
		})
		return err
	})
	if err != nil {
		return ChallengeResult{}, mapStoreError(err)
	}
	credentialID := base64.RawURLEncoding.EncodeToString(record.CredentialID)
	return ChallengeResult{OperationID: record.OperationID, ExpiresAt: record.ExpiresAt,
		PublicKey: PublicKeyRequest{Challenge: record.Challenge, RPID: record.RPID,
			AllowCredentials: []PublicKeyCredential{{Type: "public-key", ID: credentialID}},
			UserVerification: "required", Timeout: 300000}}, nil
}

func challengeContextDigest(
	environmentID, restoreOperationID uuid.UUID, binding Binding,
	rpID, origin, challenge string,
) [32]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/iam/post-restore/challenge/v1\x00"))
	for _, value := range []uuid.UUID{environmentID, restoreOperationID,
		binding.NewDatabaseSystemID, binding.TenantID, binding.OwnerUserID, binding.AnchorID} {
		_, _ = hash.Write(value[:])
	}
	_, _ = hash.Write(binding.SecurityEpoch[:])
	for _, value := range []string{rpID, origin, challenge} {
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(value)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(value))
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}
