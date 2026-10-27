package actorbridge

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestIssueAndForwardNeverOpensNetworkWhenIAMDenies(t *testing.T) {
	fixture := newRuntimeTLSFixtureForServer(t, "localhost")
	var requests atomic.Int64
	serverTLS, err := LoadServerTLSConfig(
		fixture.server, actorassertionv1.DomainForms,
		func() time.Time { return fixture.now },
	)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) { requests.Add(1) },
	))
	server.TLS = serverTLS
	server.StartTLS()
	defer server.Close()
	identity, err := LoadClientIdentity(fixture.client, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(identity,
		forwarderTargets(t, localForwarderURL(t, server.URL)),
		func() time.Time { return fixture.now })
	if err != nil {
		t.Fatal(err)
	}
	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{41}, ed25519.SeedSize))
	signer, err := actorassertionv1.NewSigner(
		"edge-2026-07", privateKey, bytes.NewReader(bytes.Repeat([]byte{9}, 32)),
	)
	if err != nil {
		t.Fatal(err)
	}
	order := &[]string{}
	issuer, err := NewIssuer(&coordinatorStub{order: order},
		authorizationStub{order: order, err: NewFault(FaultUnauthorized)},
		authorityStub{order: order}, signer, func() time.Time { return fixture.now })
	if err != nil {
		t.Fatal(err)
	}
	forwarder, err := NewForwarder(issuer, client)
	if err != nil {
		t.Fatal(err)
	}
	defer forwarder.Close()
	request, _ := http.NewRequest(http.MethodPost, "/records", bytes.NewReader([]byte("body")))
	input, err := NewForwardInput(
		actorassertionv1.DomainForms, "forms.records", "create", 1, 1, request,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = forwarder.IssueAndForward(
		context.Background(), platformtx.ScopeRequest{}, input, request,
	)
	if !FaultIs(err, FaultUnauthorized) || requests.Load() != 0 {
		t.Fatalf("forward error=%v network requests=%d", err, requests.Load())
	}
}
