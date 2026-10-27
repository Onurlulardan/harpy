package licenseauthoritysource

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	licenseauthority "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/google/uuid"
)

func TestConnectedIssuesCanonicalBoundRequestWithoutCredentialBody(t *testing.T) {
	credential := make([]byte, 32)
	if _, err := rand.Read(credential); err != nil {
		t.Fatal(err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(credential)
	request := connectedRequestFixture()
	responseDocument := []byte(`{"signed":"snapshot"}`)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, incoming *http.Request) {
		if incoming.Method != http.MethodPost || incoming.URL.Path != issuePath ||
			incoming.Header.Get("Authorization") != "Bearer "+encoded {
			t.Errorf("unexpected request: %s %s", incoming.Method, incoming.URL.Path)
		}
		body, err := io.ReadAll(incoming.Body)
		if err != nil {
			t.Error(err)
		}
		canonical, err := canonicaljsonv1.CanonicalizeObject(
			body, licenseauthority.MaximumDocumentBytes,
		)
		if err != nil || !bytes.Equal(body, canonical) || bytes.Contains(body, []byte(encoded)) {
			t.Errorf("request body is non-canonical or leaks credential: %v", err)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(responseDocument)
	}))
	defer server.Close()
	source, err := NewConnected(server.URL, append([]byte(encoded), '\n'),
		server.Client().Transport, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	document, err := source.IssueSnapshot(context.Background(), request)
	if err != nil || !bytes.Equal(document, responseDocument) {
		t.Fatalf("issue snapshot failed: %v %s", err, document)
	}
}

func TestConnectedRejectsRedirectAndInvalidCredential(t *testing.T) {
	credential := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
	targetCalled := false
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		targetCalled = true
	}))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Redirect(writer, &http.Request{}, target.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()
	source, err := NewConnected(
		redirect.URL, []byte(credential), redirect.Client().Transport, time.Second,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = source.IssueSnapshot(
		context.Background(), connectedRequestFixture(),
	); err == nil || targetCalled {
		t.Fatalf("redirect followed or accepted: called=%t err=%v", targetCalled, err)
	}
	if _, err = NewConnected(redirect.URL, []byte("short"), nil, time.Second); err == nil {
		t.Fatal("invalid refresh credential accepted")
	}
	if _, err = NewConnected(
		"http://example.com", []byte(credential), nil, time.Second,
	); err == nil {
		t.Fatal("non-loopback cleartext issuer accepted")
	}
}

func connectedRequestFixture() licenseauthority.IssueRequest {
	nonce := [32]byte{}
	copy(nonce[:], bytes.Repeat([]byte{1}, 32))
	return licenseauthority.IssueRequest{
		OperationID: uuid.New(), Nonce: nonce,
		Binding: licenseauthority.Binding{
			LicenseID: uuid.New(), LicenseeOrganizationID: uuid.New(),
			EnvironmentID: uuid.New(), DeliveryKind: licenseauthority.DeliveryConnected,
		},
		ComponentVersions: []licenseauthority.ComponentVersion{
			{Component: "core-api", Version: "v2.0.0"},
		},
	}
}
