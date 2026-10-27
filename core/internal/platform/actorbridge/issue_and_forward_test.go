package actorbridge

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type forwardedObservation struct {
	path, assertion, authorization, cookie string
}

func TestIssueAndForwardCommitsEvidenceBeforeExactMTLSRoute(t *testing.T) {
	fixture := newRuntimeTLSFixtureForServer(t, "localhost")
	observed := make(chan forwardedObservation, 1)
	serverTLS, err := LoadServerTLSConfig(
		fixture.server, actorassertionv1.DomainDocuments,
		func() time.Time { return fixture.now },
	)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			observed <- forwardedObservation{path: request.URL.RequestURI(),
				assertion:     request.Header.Get(actorassertionv1.HeaderName),
				authorization: request.Header.Get("Authorization"),
				cookie:        request.Header.Get("Cookie")}
			writer.WriteHeader(http.StatusAccepted)
		},
	))
	server.TLS = serverTLS
	server.StartTLS()
	defer server.Close()

	clientIdentity, err := LoadClientIdentity(fixture.client, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(clientIdentity,
		forwarderTargets(t, localForwarderURL(t, server.URL)),
		func() time.Time { return fixture.now })
	if err != nil {
		t.Fatal(err)
	}
	issuer, publicKey, order := forwarderIssuer(t, fixture.now)
	forwarder, err := NewForwarder(issuer, client)
	if err != nil {
		t.Fatal(err)
	}
	request, _ := http.NewRequest(
		http.MethodPost, "/api/v1/forms/records?limit=1", bytes.NewReader([]byte(`{"name":"A"}`)),
	)
	request.Header.Set("Authorization", "Bearer client-credential")
	request.Header.Set("Cookie", "session=client")
	input, err := NewForwardInput(
		actorassertionv1.DomainDocuments, "forms.records", "create", 7, 19, request,
	)
	if err != nil {
		t.Fatal(err)
	}
	response, err := forwarder.IssueAndForward(
		context.Background(), platformtx.ScopeRequest{}, input, request,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	result := <-observed
	if response.StatusCode != http.StatusAccepted ||
		result.path != "/api/v1/forms/records?limit=1" ||
		result.assertion == "" || result.authorization != "" || result.cookie != "" {
		t.Fatalf("forward result status=%d observation=%+v", response.StatusCode, result)
	}
	if len(*order) != 3 || (*order)[0] != "authorize" ||
		(*order)[1] != "authority" || (*order)[2] != "commit" {
		t.Fatalf("issuance order = %v", *order)
	}
	verifyForwardedAssertion(t, fixture.now, publicKey, input.issueInput(), result.assertion)
	forwarder.Close()
	if forwarder.Ready() {
		t.Fatal("closed forwarder remained ready")
	}
}

func forwarderTargets(t *testing.T, baseURL string) map[actorassertionv1.Domain]Target {
	t.Helper()
	targets := make(map[actorassertionv1.Domain]Target, len(exactTargetDomains))
	for _, domain := range exactTargetDomains {
		target, err := NewTarget(domain, baseURL, "localhost")
		if err != nil {
			t.Fatal(err)
		}
		targets[domain] = target
	}
	return targets
}

func localForwarderURL(t *testing.T, value string) string {
	t.Helper()
	parsed, err := url.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	_, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		t.Fatal(err)
	}
	return "https://localhost:" + port
}

func forwarderIssuer(
	t *testing.T,
	now time.Time,
) (*Issuer, ed25519.PublicKey, *[]string) {
	t.Helper()
	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{31}, ed25519.SeedSize))
	publicKey := append(ed25519.PublicKey(nil), privateKey[ed25519.SeedSize:]...)
	signer, err := actorassertionv1.NewSigner(
		"edge-2026-07", privateKey, bytes.NewReader(bytes.Repeat([]byte{8}, 32)),
	)
	if err != nil {
		t.Fatal(err)
	}
	authorization, _ := NewAuthorization(
		uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		uuid.MustParse("20000000-0000-4000-8000-000000000002"),
		uuid.MustParse("30000000-0000-4000-8000-000000000003"), 7, 19, 23,
	)
	authority, _ := NewAuthority(11, 12, 13)
	order := &[]string{}
	issuer, err := NewIssuer(&coordinatorStub{order: order},
		authorizationStub{order: order, result: authorization},
		authorityStub{order: order, result: authority}, signer,
		func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	return issuer, publicKey, order
}

func verifyForwardedAssertion(
	t *testing.T,
	now time.Time,
	publicKey ed25519.PublicKey,
	input IssueInput,
	wire string,
) {
	t.Helper()
	verifier, err := actorassertionv1.NewVerifier(
		map[string]ed25519.PublicKey{"edge-2026-07": publicKey},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = verifier.Verify(wire, actorassertionv1.Expectation{
		TenantID:     uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		TargetDomain: input.TargetDomain, ResourceKey: input.ResourceKey,
		ActionKey: input.ActionKey, TargetBindingSHA256: input.TargetBinding,
		IAMGroupGeneration: 11, IAMGeneration: 12, TargetGeneration: 13,
	}, now)
	if err != nil {
		t.Fatalf("forwarded assertion verification: %v", err)
	}
}
