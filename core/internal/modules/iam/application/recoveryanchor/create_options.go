package recoveryanchor

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) CreateOptions(
	ctx context.Context,
	scope platformtx.TenantScope,
	command anchordomain.OptionsCommand,
) (anchordomain.RegistrationOptions, error) {
	if service == nil || !scopeValid(scope, command.Subject) {
		return anchordomain.RegistrationOptions{}, anchordomain.ErrInvalid
	}
	return service.store.CreateOptions(ctx, scope, command)
}

func (runtime *Runtime) Options(
	ctx context.Context,
	input anchorapi.OptionsCommand,
) (anchorapi.RegistrationOptions, error) {
	if runtime == nil || runtime.service == nil || ctx == nil || ctx.Err() != nil {
		return anchorapi.RegistrationOptions{}, anchorapi.ErrInvalid
	}
	subject, err := parseRuntimeSubject(input.Subject)
	operationID, operationErr := uuid.Parse(input.OperationID)
	anchorID, anchorErr := uuid.Parse(input.AnchorID)
	digest, digestErr := parseRuntimeDigest(input.RequestSHA256)
	challenge, randomErr := runtimeChallenge()
	originsSHA, originsErr := runtimeOriginsSHA(runtime.origins)
	if err != nil || operationErr != nil || anchorErr != nil || digestErr != nil ||
		randomErr != nil || originsErr != nil || input.IdempotencyKey == "" {
		return anchorapi.RegistrationOptions{}, anchorapi.ErrInvalid
	}
	now := runtime.now()
	command := anchordomain.OptionsCommand{Subject: subject,
		OperationID: operationID, AnchorID: anchorID,
		IdempotencyKey: input.IdempotencyKey, RequestSHA256: digest,
		Challenge: challenge, RPID: runtime.rpID,
		Origins: append([]string(nil), runtime.origins...), OriginsSHA256: originsSHA,
		ExpiresAt: now.Add(recoveryAnchorChallengeLifetime), OccurredAt: now}
	request, err := runtime.scopeRequest(ctx, subject, "iam.recovery_anchor.options", true)
	if err != nil {
		return anchorapi.RegistrationOptions{}, err
	}
	var result anchordomain.RegistrationOptions
	err = runtime.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.service.CreateOptions(scoped, scope, command)
			return err
		})
	if err != nil {
		return anchorapi.RegistrationOptions{}, mapRuntimeError(err)
	}
	return anchorapi.RegistrationOptions{OperationID: runtimeAnchorUUID(result.OperationID),
		Challenge: result.Challenge, RPID: result.RPID,
		UserID:    base64.RawURLEncoding.EncodeToString(subject.ActorID[:]),
		ExpiresAt: result.ExpiresAt, Replayed: result.Replayed}, nil
}

func runtimeChallenge() (string, error) {
	material := make([]byte, 32)
	if _, err := rand.Read(material); err != nil {
		clear(material)
		return "", err
	}
	value := base64.RawURLEncoding.EncodeToString(material)
	clear(material)
	return value, nil
}

func runtimeOriginsSHA(origins []string) ([32]byte, error) {
	document, err := json.Marshal(origins)
	if err != nil {
		return [32]byte{}, err
	}
	digest := sha256.Sum256(document)
	clear(document)
	return digest, nil
}
