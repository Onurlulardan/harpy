package iam

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/refreshtoken"
	"github.com/google/uuid"
)

type authHTTPTestFixture struct {
	handler  *AuthHTTP
	security *HTTPSecurity
	app      *authHTTPTestApplication
	profile  accesstoken.Profile
	now      time.Time
	origin   string
	errors   []error
	claims   accesstoken.Claims
}

func newAuthHTTPTestFixture(t *testing.T) *authHTTPTestFixture {
	t.Helper()
	now := time.Unix(2_000_000_000, 0).UTC()
	origin := "http://localhost:5173"
	profile, err := accesstoken.NewProfile(
		"http://localhost:8000/auth", "http://localhost:8000/api", true,
	)
	if err != nil {
		t.Fatal(err)
	}
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))
	keyID := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{8}, 12))
	accessSigner, err := accesstoken.NewSigner(profile, keyID, private)
	if err != nil {
		t.Fatal(err)
	}
	publicKeys, err := accesstoken.NewPublicKeySet(map[string]ed25519.PublicKey{
		keyID: private.Public().(ed25519.PublicKey),
	})
	if err != nil {
		t.Fatal(err)
	}
	accessVerifier, err := accesstoken.NewVerifier(profile, publicKeys)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := csrf.NewOriginPolicy([]string{origin}, true)
	if err != nil {
		t.Fatal(err)
	}
	csrfKeyID := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{3}, 12))
	csrfKey := bytes.Repeat([]byte{9}, 32)
	csrfSigner, err := csrf.NewSigner(policy, csrfKeyID, csrfKey, rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	csrfVerifier, err := csrf.NewVerifier(policy, csrf.KeyResolverFunc(
		func(_ context.Context, requested string) ([]byte, error) {
			if requested != csrfKeyID {
				return nil, csrf.ErrInvalidKey
			}
			return append([]byte(nil), csrfKey...), nil
		},
	))
	if err != nil {
		t.Fatal(err)
	}
	security := &HTTPSecurity{
		accessProfile: profile, accessSigner: accessSigner, accessVerifier: accessVerifier,
		csrfSigner: csrfSigner, csrfVerifier: csrfVerifier,
		csrfPolicy: policy, publicKeys: publicKeys,
	}
	fixture := &authHTTPTestFixture{
		security: security, app: &authHTTPTestApplication{
			initialStartError: ErrAuthHTTPInvalidCredentials,
		},
		profile: profile, now: now, origin: origin,
	}
	handler, err := newAuthHTTP(
		security, fixture.app, func(err error) { fixture.errors = append(fixture.errors, err) },
		func() time.Time { return now },
	)
	if err != nil {
		t.Fatal(err)
	}
	fixture.handler = handler
	fixture.claims = fixture.newClaims(t, false)
	fixture.app.currentSession = fixture.sessionView(fixture.claims)
	t.Cleanup(security.Close)
	return fixture
}

func (fixture *authHTTPTestFixture) newClaims(
	t *testing.T,
	portal bool,
) accesstoken.Claims {
	t.Helper()
	input := accesstoken.ClaimsInput{
		Subject: uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		Session: uuid.MustParse("20000000-0000-4000-8000-000000000001"),
		Tenant:  uuid.MustParse("30000000-0000-4000-8000-000000000001"),
		Client:  accesstoken.ClientOperatorConsole, IssuedAt: fixture.now,
		JTI:          uuid.MustParse("40000000-0000-4000-8000-000000000001"),
		TokenVersion: 3, AuthorizationVersion: 5,
		PlatformAuthorizationClock: 7, SecurityEpoch: [32]byte{1},
		AccessMode:         accesstoken.AccessModeOperator,
		AuthenticationTime: fixture.now.Add(-time.Minute),
		Assurance:          accesstoken.AssuranceBaseline,
		Methods:            []accesstoken.AuthMethod{accesstoken.AuthMethodPassword},
	}
	if portal {
		binding := uuid.MustParse("50000000-0000-4000-8000-000000000001")
		input.Client, input.AccessMode = accesstoken.ClientPortal, accesstoken.AccessModePortal
		input.PortalBinding = &binding
	}
	claims, err := accesstoken.NewClaims(fixture.profile, input)
	if err != nil {
		t.Fatal(err)
	}
	return claims
}

func (fixture *authHTTPTestFixture) newGrant(
	t *testing.T,
	claims accesstoken.Claims,
) *authHTTPTestGrant {
	t.Helper()
	access, err := fixture.security.accessSigner.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := refreshtoken.NewGenerator().Generate()
	if err != nil {
		t.Fatal(err)
	}
	return &authHTTPTestGrant{
		access: access, refresh: refresh,
		absolute: fixture.now.Add(90 * 24 * time.Hour),
	}
}
