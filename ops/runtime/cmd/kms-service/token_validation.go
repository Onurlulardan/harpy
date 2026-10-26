package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type tokenClaims struct {
	Audience  string `json:"aud"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	Issuer    string `json:"iss"`
	ID        string `json:"jti"`
	NotBefore int64  `json:"nbf,omitempty"`
	Subject   string `json:"sub"`
}

type subjectTokenHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Type      string `json:"typ"`
}

func (authority *tokenAuthority) validateSubject(token string) (string, error) {
	parts, err := compactTokenParts(token)
	if err != nil {
		return "", err
	}
	var header subjectTokenHeader
	if decodeTokenJSON(parts[0], &header) != nil ||
		header.Algorithm != "EdDSA" || header.Type != "JWT" {
		return "", fmt.Errorf("subject token header is invalid")
	}
	workload, exists := authority.workloads[header.KeyID]
	signature, signatureErr := base64.RawURLEncoding.DecodeString(parts[2])
	if !exists || signatureErr != nil || !ed25519.Verify(
		workload.publicKey, []byte(parts[0]+"."+parts[1]), signature,
	) {
		return "", fmt.Errorf("subject token signature is invalid")
	}
	var claims tokenClaims
	now := authority.clock().UTC().Unix()
	if decodeTokenJSON(parts[1], &claims) != nil ||
		claims.Audience != hostAudience || claims.Issuer != "harpy-host" ||
		claims.Subject != workload.subject || claims.NotBefore != claims.IssuedAt-30 ||
		claims.NotBefore > now || claims.IssuedAt > now+30 ||
		claims.ExpiresAt <= now || claims.ExpiresAt-claims.IssuedAt != 300 ||
		claims.ID == "" {
		return "", fmt.Errorf("subject token is invalid")
	}
	return claims.Subject, nil
}

func (authority *tokenAuthority) validateAccess(token string) error {
	parts, err := compactTokenParts(token)
	if err != nil {
		return err
	}
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || string(header) != `{"alg":"HS256","typ":"JWT"}` {
		return fmt.Errorf("access token header is invalid")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	signer := hmac.New(sha256.New, authority.key[:])
	_, _ = signer.Write([]byte(parts[0] + "." + parts[1]))
	if err != nil || !hmac.Equal(signature, signer.Sum(nil)) {
		return fmt.Errorf("access token signature is invalid")
	}
	var claims tokenClaims
	now := authority.clock().UTC().Unix()
	if decodeTokenJSON(parts[1], &claims) != nil ||
		claims.Audience != accessAudience || claims.Issuer != "harpy-host" ||
		!authority.knownSubject(claims.Subject) ||
		claims.IssuedAt > now+30 || claims.ExpiresAt <= now ||
		claims.ExpiresAt-claims.IssuedAt != 300 || claims.NotBefore != 0 ||
		claims.ID == "" {
		return fmt.Errorf("access token is invalid")
	}
	return nil
}

func (authority *tokenAuthority) knownSubject(subject string) bool {
	for _, workload := range authority.workloads {
		if workload.subject == subject {
			return true
		}
	}
	return false
}

func compactTokenParts(token string) ([]string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("token is invalid")
	}
	return parts, nil
}

func decodeTokenJSON(encoded string, target any) error {
	document, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("token segment is invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(target) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return fmt.Errorf("token JSON is invalid")
	}
	return nil
}
