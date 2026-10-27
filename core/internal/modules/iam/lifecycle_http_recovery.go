package iam

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const lifecycleRecoveryCodeCount = 10

func (application *lifecycleHTTPApplication) IssueRecoveryCodes(
	ctx context.Context, claims accesstoken.Claims, input LifecycleRecoveryIssueInput,
) (LifecycleRecoveryIssueView, error) {
	if input.DisplayLabel == "" || !identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return LifecycleRecoveryIssueView{}, ErrLifecycleHTTPInvalid
	}
	keyID, key, err := application.recovery.Active()
	if err != nil {
		return LifecycleRecoveryIssueView{}, ErrLifecycleHTTPUnavailable
	}
	defer clear(key[:])
	operationID := lifecycleOperationID(claims, "iam.recovery_codes.issue", input.IdempotencyKey)
	authenticatorID := lifecycleChildID(operationID, "authenticator")
	codes, wires, err := generateLifecycleRecoveryCodes(operationID, keyID, key)
	if err != nil {
		return LifecycleRecoveryIssueView{}, err
	}
	requestSHA, _ := lifecycleRequestSHA(struct{ Label string }{input.DisplayLabel})
	var result lifecycle.MutationResult
	err = application.inCommand(ctx, claims, "iam.recovery_codes.issue", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var executeErr error
			result, executeErr = application.runtime.service.RotateRecoveryCodes(scoped, scope,
				lifecycle.RecoveryIssueCommand{
					Evidence:        lifecycleEvidence(claims, permissionID, operationID, requestSHA),
					AuthenticatorID: authenticatorID, UserID: claims.Subject(),
					DisplayLabel: input.DisplayLabel, HMACKeyID: keyID, Codes: codes,
				})
			return executeErr
		})
	if err != nil {
		clearRecoveryWires(wires)
		return LifecycleRecoveryIssueView{}, err
	}
	if result.Replayed {
		clearRecoveryWires(wires)
		return LifecycleRecoveryIssueView{}, ErrLifecycleHTTPConflict
	}
	return LifecycleRecoveryIssueView{AuthenticatorID: authenticatorID, Codes: wires}, nil
}

func generateLifecycleRecoveryCodes(
	operationID uuid.UUID, keyID string, key [32]byte,
) ([]lifecycle.RecoveryCodeIssue, []string, error) {
	codes := make([]lifecycle.RecoveryCodeIssue, lifecycleRecoveryCodeCount)
	wires := make([]string, lifecycleRecoveryCodeCount)
	for index := range codes {
		secret := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, secret); err != nil {
			clear(secret)
			clearRecoveryWires(wires)
			return nil, nil, ErrLifecycleHTTPUnavailable
		}
		mac := hmac.New(sha256.New, key[:])
		_, _ = mac.Write([]byte("harpy:iam-recovery-code:v1\x00"))
		_, _ = mac.Write(secret)
		digest := mac.Sum(nil)
		codes[index] = lifecycle.RecoveryCodeIssue{ID: lifecycleChildID(operationID,
			"recovery-code-"+string(rune('a'+index))), HMACSHA256: hex.EncodeToString(digest)}
		wires[index] = keyID + "." + base64.RawURLEncoding.EncodeToString(secret)
		clear(secret)
		clear(digest)
	}
	return codes, wires, nil
}

func clearRecoveryWires(values []string) {
	for index := range values {
		values[index] = ""
	}
}

func (application *lifecycleHTTPApplication) ListRecoveryCodes(
	ctx context.Context, claims accesstoken.Claims,
) ([]lifecycle.RecoveryCode, error) {
	var result []lifecycle.RecoveryCode
	err := application.inSnapshot(ctx, claims, "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var readErr error
			result, readErr = application.runtime.service.ReadRecoveryCodes(scoped, scope,
				lifecycle.RecoveryListQuery{UserID: claims.Subject(), ActorID: claims.Subject(),
					PermissionID: permissionID, ExpectedTokenVersion: claims.TokenVersion(),
					ExpectedAuthorizationVersion: claims.AuthorizationVersion()})
			return readErr
		})
	return result, err
}
