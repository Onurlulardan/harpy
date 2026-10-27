package aiprovidercontextv1

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

const (
	ContractVersion   = "core-ai-provider-context.v1"
	Issuer            = "urn:harpy:core-api"
	Audience          = "harpy-ai-provider-v1"
	MaximumTokenBytes = 8 << 10
	MaximumLifetime   = 60 * time.Second
)

var ErrInvalidContext = errors.New("AI provider context is invalid")

type Claims struct {
	ContractVersion      string `json:"contract_version"`
	Issuer               string `json:"issuer"`
	Audience             string `json:"audience"`
	KeyID                string `json:"key_id"`
	TenantID             string `json:"tenant_id"`
	ActorID              string `json:"actor_id"`
	ConfigID             string `json:"config_id"`
	ConfigRevision       int64  `json:"config_revision"`
	AuthorizationVersion int64  `json:"authorization_version"`
	Permission           string `json:"permission"`
	Operation            string `json:"operation"`
	CorrelationID        string `json:"correlation_id"`
	JTI                  string `json:"jti"`
	IssuedAt             int64  `json:"issued_at"`
	NotBefore            int64  `json:"not_before"`
	ExpiresAt            int64  `json:"expires_at"`
}

type Token struct {
	Claims    Claims `json:"claims"`
	Signature string `json:"signature"`
}

type Expected struct {
	TenantID, ConfigID, Operation        string
	ConfigRevision, AuthorizationVersion int64
	Now                                  time.Time
}

func claimsBytes(claims Claims) ([]byte, error) {
	document, err := json.Marshal(claims)
	if err != nil {
		return nil, ErrInvalidContext
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 4096)
	if err != nil {
		return nil, ErrInvalidContext
	}
	return canonical, nil
}

func (claims Claims) valid() bool {
	for _, value := range []string{claims.TenantID, claims.ActorID, claims.ConfigID,
		claims.CorrelationID, claims.JTI} {
		if _, err := uuid.Parse(value); err != nil {
			return false
		}
	}
	lifetime := time.Duration(claims.ExpiresAt-claims.IssuedAt) * time.Second
	return claims.ContractVersion == ContractVersion && claims.Issuer == Issuer &&
		claims.Audience == Audience && claims.KeyID != "" && claims.ConfigRevision > 0 &&
		claims.AuthorizationVersion > 0 && claims.Permission == "ai.config.manage" &&
		claims.Operation != "" && claims.NotBefore >= claims.IssuedAt &&
		lifetime > 0 && lifetime <= MaximumLifetime
}

func canonicalToken(token Token) ([]byte, error) {
	document, err := json.Marshal(token)
	if err != nil {
		return nil, ErrInvalidContext
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumTokenBytes)
	if err != nil {
		return nil, ErrInvalidContext
	}
	return canonical, nil
}

func Parse(document []byte) (Token, error) {
	if len(document) == 0 || len(document) > MaximumTokenBytes {
		return Token{}, ErrInvalidContext
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumTokenBytes)
	if err != nil || !bytes.Equal(canonical, document) {
		return Token{}, ErrInvalidContext
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var token Token
	if err := decoder.Decode(&token); err != nil || !token.Claims.valid() {
		return Token{}, ErrInvalidContext
	}
	signature, err := base64.RawURLEncoding.DecodeString(token.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize ||
		base64.RawURLEncoding.EncodeToString(signature) != token.Signature {
		return Token{}, ErrInvalidContext
	}
	return token, nil
}
