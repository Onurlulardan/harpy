package aiprovidercontextv2

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	aiinferencev1 "github.com/Onurlulardan/harpy/contracts/aiinference/v1"
	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

const (
	ContractVersion   = "core-ai-provider-context.v2"
	Issuer            = "urn:harpy:core-api"
	Audience          = aiinferencev1.ServiceAudience
	MaximumTokenBytes = 12 << 10
	MaximumLifetime   = 60 * time.Second
)

var (
	ErrInvalidContext = errors.New("AI provider context v2 is invalid")
	hexPattern        = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type Claims struct {
	ContractVersion        string `json:"contract_version"`
	Issuer                 string `json:"issuer"`
	Audience               string `json:"audience"`
	KeyID                  string `json:"key_id"`
	TenantID               string `json:"tenant_id"`
	ActorID                string `json:"actor_id"`
	ConfigID               string `json:"config_id"`
	ConfigRevision         int64  `json:"config_revision"`
	CredentialRevision     int64  `json:"credential_revision"`
	AuthorizationVersion   int64  `json:"authorization_version"`
	Permission             string `json:"permission"`
	Operation              string `json:"operation"`
	InvocationID           string `json:"invocation_id"`
	ReceiptID              string `json:"receipt_id"`
	RequestHash            string `json:"request_hash"`
	CredentialEnvelopeHash string `json:"credential_envelope_sha256,omitempty"`
	CorrelationID          string `json:"correlation_id"`
	JTI                    string `json:"jti"`
	IssuedAt               int64  `json:"issued_at"`
	NotBefore              int64  `json:"not_before"`
	ExpiresAt              int64  `json:"expires_at"`
}

type Token struct {
	Claims    Claims `json:"claims"`
	Signature string `json:"signature"`
}

type Expected struct {
	TenantID, ConfigID, Operation        string
	InvocationID, ReceiptID, RequestHash string
	CredentialEnvelopeHash               string
	ConfigRevision, CredentialRevision   int64
	AuthorizationVersion                 int64
	Now, HardDeadline                    time.Time
}

func NewClaims(now time.Time, lifetime time.Duration) Claims {
	issued := now.UTC().Truncate(time.Second)
	return Claims{IssuedAt: issued.Unix(), NotBefore: issued.Unix(),
		ExpiresAt: issued.Add(lifetime).Unix()}
}

func claimsBytes(claims Claims) ([]byte, error) {
	document, err := json.Marshal(claims)
	if err != nil {
		return nil, ErrInvalidContext
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumTokenBytes)
	clear(document)
	if err != nil {
		return nil, ErrInvalidContext
	}
	return canonical, nil
}

func canonicalToken(token Token) ([]byte, error) {
	document, err := json.Marshal(token)
	if err != nil {
		return nil, ErrInvalidContext
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumTokenBytes)
	clear(document)
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

func (claims Claims) valid() bool {
	invocation, invocationErr := uuid.Parse(claims.InvocationID)
	receipt, receiptErr := uuid.Parse(claims.ReceiptID)
	for _, value := range []string{claims.TenantID, claims.ActorID, claims.ConfigID,
		claims.CorrelationID, claims.JTI} {
		if _, err := uuid.Parse(value); err != nil {
			return false
		}
	}
	lifetime := time.Duration(claims.ExpiresAt-claims.IssuedAt) * time.Second
	return claims.ContractVersion == ContractVersion && claims.Issuer == Issuer &&
		claims.Audience == Audience && claims.KeyID != "" && claims.ConfigRevision > 0 &&
		claims.CredentialRevision >= 0 && claims.AuthorizationVersion > 0 &&
		claims.Permission == aiinferencev1.PermissionExecute && validOperation(claims.Operation) &&
		invocationErr == nil && receiptErr == nil && invocation.Version() == 7 && receipt.Version() == 7 &&
		hexPattern.MatchString(claims.RequestHash) && validEnvelopeHash(claims) &&
		claims.NotBefore >= claims.IssuedAt && lifetime > 0 && lifetime <= MaximumLifetime
}

func validEnvelopeHash(claims Claims) bool {
	return (claims.CredentialRevision == 0 && claims.CredentialEnvelopeHash == "") ||
		(claims.CredentialRevision > 0 && hexPattern.MatchString(claims.CredentialEnvelopeHash))
}

func validOperation(value string) bool {
	return value == aiinferencev1.OperationGenerate || value == aiinferencev1.OperationStream ||
		value == aiinferencev1.OperationEmbed
}
