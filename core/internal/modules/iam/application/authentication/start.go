package authentication

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (authenticator Authenticator) StartReauthentication(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	request BeginReauthenticationRequest,
) (ReauthenticationChallenge, error) {
	if !authenticator.valid || request.validate() != nil ||
		!validCredentialScope(
			ctx, scope, platformtx.CredentialReauthentication,
			platformtx.CredentialSessionLocator,
		) || !request.matchesScope(scope) {
		return ReauthenticationChallenge{}, ErrAuthenticationContract
	}
	material, err := authenticator.security.Challenges.Generate()
	if err != nil || nilInterface(material) {
		return ReauthenticationChallenge{}, ErrAuthenticationCrypto
	}
	destroyMaterial := true
	defer func() {
		if destroyMaterial {
			material.Destroy()
		}
	}()
	keyID, key, err := authenticator.security.ChallengeKeys.Active()
	if err != nil {
		return ReauthenticationChallenge{}, ErrAuthenticationCrypto
	}
	challengeHMAC, err := material.HMAC(key[:])
	clear(key[:])
	if err != nil {
		return ReauthenticationChallenge{}, ErrAuthenticationCrypto
	}
	defer clear(challengeHMAC[:])
	challengeID, err := authenticator.identifiers()
	if err != nil || challengeID == uuid.Nil {
		return ReauthenticationChallenge{}, ErrAuthenticationCrypto
	}
	contextSHA, err := request.contextSHA256(scope.TenantID())
	if err != nil {
		return ReauthenticationChallenge{}, err
	}
	created, err := authenticator.port.StartReauthentication(
		ctx, scope, request.creation(challengeID, keyID, challengeHMAC, contextSHA),
	)
	if err != nil || created.validate(request.Method) != nil {
		return ReauthenticationChallenge{}, ErrAuthenticationStorage
	}
	destroyMaterial = false
	return ReauthenticationChallenge{
		ID: challengeID, Method: request.Method, ExpiresAt: created.ExpiresAt.UTC(),
		CredentialID:             append([]byte(nil), created.CredentialID...),
		RPID:                     created.RPID,
		UserVerificationRequired: created.UserVerificationRequired,
		material:                 material, valid: true,
	}, nil
}
