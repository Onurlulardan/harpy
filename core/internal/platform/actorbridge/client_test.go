package actorbridge

import (
	"net/http"
	"strings"
	"testing"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func TestClientPreparesOnlyExactMTLSTargetAndStripsCredentials(t *testing.T) {
	fixture := newRuntimeTLSFixture(t)
	identity, err := LoadClientIdentity(fixture.client, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(identity, runtimeTargets(t), func() time.Time { return fixture.now })
	if err != nil {
		t.Fatal(err)
	}
	request, _ := http.NewRequest(http.MethodPost, "/api/v1/forms/records?limit=1", nil)
	request.Header = http.Header{
		"authorization":                   []string{"Bearer client-token"},
		"Cookie":                          []string{"session=client"},
		"harpy-actor-assertion":       []string{"attacker"},
		"Connection":                      []string{"X-Remove"},
		"X-Remove":                        []string{"secret"},
		"X-Forwarded-Untrusted-Dimension": []string{"attacker"},
		"X-Tenant-ID":                     []string{"10000000-0000-4000-8000-000000000001"},
	}
	credential := runtimeCredential(t, fixture.now)
	prepared, err := client.Prepare(request, actorassertionv1.DomainDocuments, credential)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.URL.String() != "https://v1-documents.fixture.invalid/api/v1/forms/records?limit=1" ||
		!hasHeaderFold(prepared.Header, actorassertionv1.HeaderName) ||
		!hasHeaderFold(prepared.Header, "X-Tenant-ID") {
		t.Fatalf("prepared request differs: %s %#v", prepared.URL, prepared.Header)
	}
	for _, prohibited := range []string{
		"Authorization", "Cookie", "X-Remove", "X-Forwarded-Untrusted-Dimension",
	} {
		if hasHeaderFold(prepared.Header, prohibited) {
			t.Fatalf("prepared request retained %s", prohibited)
		}
	}
	transport := client.targets[actorassertionv1.DomainDocuments]
	if transport.transport.Proxy != nil || transport.client.CheckRedirect == nil {
		t.Fatal("client allows environment proxies or redirects")
	}
	if err := transport.client.CheckRedirect(prepared, []*http.Request{request}); err == nil {
		t.Fatal("actor assertion redirect was accepted")
	}
	client.Close()
	if client.Ready() {
		t.Fatal("closed client remained ready")
	}
}

func TestClientRequiresExactTargetRegistry(t *testing.T) {
	fixture := newRuntimeTLSFixture(t)
	identity, _ := LoadClientIdentity(fixture.client, fixture.now)
	targets := runtimeTargets(t)
	delete(targets, actorassertionv1.DomainWork)
	if _, err := NewClient(identity, targets, time.Now); !FaultIs(err, FaultInvalid) {
		t.Fatalf("incomplete target registry error = %v", err)
	}
	if _, err := NewTarget(
		actorassertionv1.Domain("iam"), "https://iam.fixture.invalid", "iam.fixture.invalid",
	); !FaultIs(err, FaultInvalid) {
		t.Fatalf("IAM target error = %v", err)
	}
	if _, err := NewTarget(
		actorassertionv1.DomainDocuments, "http://v1-forms.fixture.invalid",
		"v1-forms.fixture.invalid",
	); !FaultIs(err, FaultInvalid) || strings.Contains(err.Error(), "http://") {
		t.Fatalf("plaintext target error = %v", err)
	}
}
