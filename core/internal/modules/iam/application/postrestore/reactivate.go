package postrestore

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

func (service *Service) Reactivate(
	ctx context.Context, environmentID, restoreOperationID, operationID uuid.UUID,
	origin string, assertionDocument []byte, newPassword []byte,
) (ReactivateResult, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || environmentID == uuid.Nil ||
		restoreOperationID == uuid.Nil || operationID == uuid.Nil || !service.validOrigin(origin) ||
		len(assertionDocument) < 2 || len(assertionDocument) > 64*1024 {
		return ReactivateResult{}, ErrDenied
	}
	now := service.clock().UTC()
	challenge, err := service.store.ChallengeContext(
		ctx, operationID, environmentID, restoreOperationID, origin, now,
	)
	if err != nil {
		return ReactivateResult{}, ErrDenied
	}
	binding, err := service.store.Binding(ctx, environmentID, restoreOperationID)
	expectedContext := challengeContextDigest(environmentID, restoreOperationID,
		binding, challenge.RPID, origin, challenge.Challenge)
	if err != nil || subtle.ConstantTimeCompare(
		expectedContext[:], challenge.ContextSHA256[:],
	) != 1 {
		return ReactivateResult{}, ErrDenied
	}
	result, err := service.security.Assertions.Verify(assertionDocument, AssertionExpectation{
		Challenge: []byte(challenge.Challenge), CredentialID: challenge.CredentialID,
		PublicKeyCOSE: challenge.PublicKeyCOSE, RPID: challenge.RPID, Origin: origin,
		UserHandle: challenge.OwnerUserID[:], UserVerificationRequired: true,
	})
	if err != nil || !result.UserVerified || result.BackupEligible || result.BackupState ||
		(challenge.SignCount > 0 && int64(result.SignCount) <= challenge.SignCount) {
		service.fail(ctx, operationID, environmentID, restoreOperationID, now)
		return ReactivateResult{}, ErrDenied
	}
	password, err := service.security.Passwords.Issue(
		ctx, newPassword, []string{challenge.Identifier.NormalizedValue},
	)
	if err != nil {
		return ReactivateResult{}, ErrInvalid
	}
	assertionDigest := sha256.Sum256(assertionDocument)
	passwordDigest := passwordCredentialDigest(password)
	receiptDigest := reactivationReceiptDigest(operationID, restoreOperationID,
		challenge.TenantID, challenge.OwnerUserID, challenge.AnchorID,
		assertionDigest, passwordDigest, int64(result.SignCount))
	err = service.store.PersistReactivation(ctx, CompleteCommand{
		OperationID: operationID, EnvironmentID: environmentID,
		RestoreOperationID: restoreOperationID, TenantID: challenge.TenantID,
		OwnerUserID: challenge.OwnerUserID, AnchorID: challenge.AnchorID,
		AssertionSHA256: assertionDigest, PasswordSHA256: passwordDigest,
		ReceiptSHA256: receiptDigest, SignCount: int64(result.SignCount),
		Password: password, Identifier: challenge.Identifier,
		DirectoryKeyID: service.directory.ID,
		DirectoryHMAC:  service.directoryHMAC(challenge.Identifier.NormalizedValue),
		OccurredAt:     now,
	})
	if err != nil {
		return ReactivateResult{}, mapStoreError(err)
	}
	return ReactivateResult{UserID: challenge.OwnerUserID,
		TenantID: challenge.TenantID, TenantSlug: binding.TenantSlug,
		Identifier: challenge.Identifier.NormalizedValue, SessionState: "active"}, nil
}

func (service *Service) fail(
	ctx context.Context, operationID, environmentID, restoreOperationID uuid.UUID, now time.Time,
) {
	_ = service.store.FailChallenge(ctx, operationID, environmentID, restoreOperationID, now)
}

func passwordCredentialDigest(credential PasswordCredential) [32]byte {
	document, _ := json.Marshal(struct {
		Algorithm   string `json:"algorithm"`
		Salt        []byte `json:"salt"`
		Tag         []byte `json:"tag"`
		Memory      uint32 `json:"memory_kib"`
		Iterations  uint32 `json:"iterations"`
		Parallelism uint8  `json:"parallelism"`
	}{"argon2id_v1", credential.Salt[:], credential.Tag[:], credential.MemoryKiB,
		credential.Iterations, credential.Parallelism})
	return sha256.Sum256(document)
}

func reactivationReceiptDigest(
	operationID, restoreID, tenantID, ownerID, anchorID uuid.UUID,
	assertion, password [32]byte, signCount int64,
) [32]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/iam/post-restore/reactivation-receipt/v1\x00"))
	for _, value := range []uuid.UUID{operationID, restoreID, tenantID, ownerID, anchorID} {
		_, _ = hash.Write(value[:])
	}
	_, _ = hash.Write(assertion[:])
	_, _ = hash.Write(password[:])
	var count [8]byte
	binary.BigEndian.PutUint64(count[:], uint64(signCount))
	_, _ = hash.Write(count[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}
