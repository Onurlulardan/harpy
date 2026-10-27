package iam

import (
	"crypto/ed25519"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type HTTPSecurity struct {
	accessProfile  accesstoken.Profile
	accessSigner   *accesstoken.Signer
	accessVerifier *accesstoken.Verifier
	csrfSigner     *csrf.Signer
	csrfVerifier   *csrf.Verifier
	csrfPolicy     csrf.OriginPolicy
	publicKeys     *accesstoken.PublicKeySet
}

func NewHTTPSecurity(iamConfig config.IAMAPI) (*HTTPSecurity, error) {
	httpProfile, err := iamConfig.HTTPProfile()
	if err != nil {
		return nil, fmt.Errorf("load IAM HTTP profile: %w", err)
	}
	accessProfile, err := accesstoken.NewProfile(
		httpProfile.Issuer(), httpProfile.Audience(), httpProfile.AllowLocalHTTP(),
	)
	if err != nil {
		return nil, fmt.Errorf("build access-token profile: %w", err)
	}
	accessKeys, err := iamConfig.AccessTokenKeyring()
	if err != nil {
		return nil, fmt.Errorf("load access-token keyring: %w", err)
	}
	keyID, privateKey, err := accessKeys.SigningKey()
	if err != nil {
		return nil, fmt.Errorf("load access-token signing key: %w", err)
	}
	defer clear(privateKey)
	accessSigner, err := accesstoken.NewSigner(accessProfile, keyID, privateKey)
	if err != nil {
		return nil, fmt.Errorf("build access-token signer: %w", err)
	}
	publicKeyMaterial, err := accessKeys.PublicKeys()
	if err != nil {
		accessSigner.Close()
		return nil, fmt.Errorf("load access-token public keys: %w", err)
	}
	publicKeys, err := accesstoken.NewPublicKeySet(publicKeyMaterial)
	clearPublicKeys(publicKeyMaterial)
	if err != nil {
		accessSigner.Close()
		return nil, fmt.Errorf("build access-token public key set: %w", err)
	}
	accessVerifier, err := accesstoken.NewVerifier(accessProfile, publicKeys)
	if err != nil {
		accessSigner.Close()
		return nil, fmt.Errorf("build access-token verifier: %w", err)
	}
	security, err := buildCSRFSecurity(iamConfig, httpProfile, accessSigner)
	if err != nil {
		return nil, err
	}
	security.accessVerifier = accessVerifier
	security.accessProfile = accessProfile
	security.publicKeys = publicKeys
	return security, nil
}

func buildCSRFSecurity(
	iamConfig config.IAMAPI,
	httpProfile config.IAMHTTPProfile,
	accessSigner *accesstoken.Signer,
) (*HTTPSecurity, error) {
	policy, err := csrf.NewOriginPolicy(
		httpProfile.AllowedOrigins(), httpProfile.AllowLocalHTTP(),
	)
	if err != nil {
		accessSigner.Close()
		return nil, fmt.Errorf("build CSRF origin policy: %w", err)
	}
	keyring, err := iamConfig.CSRFKeyring()
	if err != nil {
		accessSigner.Close()
		return nil, fmt.Errorf("load CSRF keyring: %w", err)
	}
	keyID, material, err := keyring.Active()
	if err != nil {
		accessSigner.Close()
		return nil, fmt.Errorf("load CSRF signing key: %w", err)
	}
	defer clear(material[:])
	csrfSigner, err := csrf.NewProductionSigner(policy, keyID, material[:])
	if err != nil {
		accessSigner.Close()
		return nil, fmt.Errorf("build CSRF signer: %w", err)
	}
	csrfVerifier, err := csrf.NewVerifier(policy, keyring)
	if err != nil {
		csrfSigner.Close()
		accessSigner.Close()
		return nil, fmt.Errorf("build CSRF verifier: %w", err)
	}
	return &HTTPSecurity{
		accessSigner: accessSigner, csrfSigner: csrfSigner,
		csrfVerifier: csrfVerifier, csrfPolicy: policy,
	}, nil
}

func (security *HTTPSecurity) NewJWKSHandler(
	report HTTPErrorSink,
) (*JWKSHandler, error) {
	if !security.Valid() {
		return nil, fmt.Errorf("IAM HTTP security is incomplete")
	}
	return NewJWKSHandler(security.publicKeys, report)
}

func (security *HTTPSecurity) Valid() bool {
	return security != nil && security.accessSigner != nil &&
		security.accessProfile.Issuer() != "" &&
		security.accessVerifier != nil && security.csrfSigner != nil &&
		security.csrfVerifier != nil && security.publicKeys != nil
}

func (security *HTTPSecurity) browserOriginAllowed(origin string) bool {
	return security != nil && security.csrfPolicy.Allows(origin)
}

func (security *HTTPSecurity) Close() {
	if security == nil {
		return
	}
	security.accessSigner.Close()
	security.csrfSigner.Close()
}

func clearPublicKeys(keys map[string]ed25519.PublicKey) {
	for keyID, key := range keys {
		clear(key)
		delete(keys, keyID)
	}
}
