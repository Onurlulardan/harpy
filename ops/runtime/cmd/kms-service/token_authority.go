package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	hostAudience = "//iam.googleapis.com/projects/123456789012/locations/global/" +
		"workloadIdentityPools/harpy-host/providers/host-runtime"
	accessAudience = "harpy-host-kms"
)

type tokenAuthority struct {
	profile   string
	key       [32]byte
	workloads map[string]workloadIdentity
	clock     func() time.Time
}

func localTokenAuthority() *tokenAuthority {
	return newTokenAuthority("local", nil, nil, time.Now)
}

func newTokenAuthority(
	profile string,
	key []byte,
	workloads map[string]workloadIdentity,
	clock func() time.Time,
) *tokenAuthority {
	authority := &tokenAuthority{
		profile: profile, workloads: workloads, clock: clock,
	}
	copy(authority.key[:], key)
	return authority
}

func (authority *tokenAuthority) exchange(form url.Values) (string, int, error) {
	subject, err := authority.validateSubject(form.Get("subject_token"))
	if err != nil || form.Get("audience") != hostAudience {
		return "", 0, fmt.Errorf("subject token is invalid")
	}
	now := authority.clock().UTC()
	token, err := authority.sign(map[string]any{
		"aud": accessAudience, "exp": now.Add(5 * time.Minute).Unix(),
		"iat": now.Unix(), "iss": "harpy-host", "jti": randomID(),
		"sub": subject,
	})
	return token, 300, err
}

func (authority *tokenAuthority) authorize(
	ctx context.Context,
	request any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	values := metadata.ValueFromIncomingContext(ctx, "authorization")
	if len(values) != 1 || !strings.HasPrefix(values[0], "Bearer ") ||
		authority.validateAccess(strings.TrimPrefix(values[0], "Bearer ")) != nil {
		return nil, status.Error(codes.Unauthenticated, "KMS authentication failed")
	}
	return handler(ctx, request)
}

func randomID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(value[:])
}

func (authority *tokenAuthority) sign(payload map[string]any) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil || payload["jti"] == "" {
		return "", fmt.Errorf("token signing failed")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("token signing failed")
	}
	wire := base64.RawURLEncoding.EncodeToString(header) + "." +
		base64.RawURLEncoding.EncodeToString(body)
	signer := hmac.New(sha256.New, authority.key[:])
	_, _ = signer.Write([]byte(wire))
	return wire + "." + base64.RawURLEncoding.EncodeToString(signer.Sum(nil)), nil
}
