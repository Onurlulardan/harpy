package asyncmanifestv1

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestReleaseManifestContainsInstallationActivationProducer(t *testing.T) {
	manifest, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	policy, err := manifest.Producers().Resolve("installation.activated", 1)
	if err != nil {
		t.Fatal(err)
	}
	if policy.OwnerDomain != "iam" ||
		policy.PayloadSchemaSHA256 != "4a44e2f3f9e2669280d713512812010d50bf9f077f4ee1a0709e80c3633dcb5e" {
		t.Fatalf("installation activation producer = %#v", policy)
	}
}

func TestReleaseManifestContainsIdentityChangeProducer(t *testing.T) {
	manifest, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	policy, err := manifest.Producers().Resolve("iam.identity.changed", 1)
	if err != nil {
		t.Fatal(err)
	}
	document, err := os.ReadFile("../../events/v1/iam-identity-change-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(document))
	if policy.OwnerDomain != "iam" || policy.PayloadSchemaSHA256 != digest {
		t.Fatalf("identity change producer = %#v", policy)
	}
}

func TestReleaseManifestContainsPermissionRequestDecisionProducer(t *testing.T) {
	manifest, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	policy, err := manifest.Producers().Resolve("iam.permission_request.decided", 1)
	if err != nil {
		t.Fatal(err)
	}
	document, err := os.ReadFile("../../events/v1/iam-permission-request-decided-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(document))
	if policy.OwnerDomain != "iam" || policy.PayloadSchemaSHA256 != digest {
		t.Fatalf("permission request producer = %#v", policy)
	}
}

func TestReleaseManifestContainsPermissionRequestTransitionProducer(t *testing.T) {
	manifest, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	policy, err := manifest.Producers().Resolve("iam.permission_request.transitioned", 1)
	if err != nil {
		t.Fatal(err)
	}
	document, err := os.ReadFile("../../events/v1/iam-permission-request-transitioned-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(document))
	if policy.OwnerDomain != "iam" || policy.PayloadSchemaSHA256 != digest {
		t.Fatalf("permission request transition producer = %#v", policy)
	}
}

func TestReleaseManifestIsDeterministic(t *testing.T) {
	first, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	second, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	assertSameDigest(t, first.Jobs().SHA256, second.Jobs().SHA256, "jobs")
	assertSameDigest(t, first.Producers().SHA256, second.Producers().SHA256, "producers")
	assertSameDigest(t, first.Consumers().SHA256, second.Consumers().SHA256, "consumers")
}

func TestReleaseManifestContainsExactWorkerPolicies(t *testing.T) {
	manifest, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	policies := manifest.Jobs().Policies()
	type expectedPolicy struct {
		owner    string
		effect   jobsv1.EffectClass
		attempts int
	}
	expected := map[string]expectedPolicy{}
	for _, policy := range jobPolicies() {
		expected[policy.Kind] = expectedPolicy{
			owner: policy.OwnerDomain, effect: policy.EffectClass,
			attempts: policy.MaximumAttempts,
		}
	}
	if len(policies) != len(expected) {
		t.Fatalf("job policy count = %d", len(policies))
	}
	for index, policy := range policies {
		expectedValue, exists := expected[policy.Kind]
		if !exists || policy.PayloadRevision != 1 ||
			policy.OwnerDomain != expectedValue.owner ||
			policy.EffectClass != expectedValue.effect ||
			policy.MaximumAttempts != expectedValue.attempts {
			t.Fatalf("job policy %d = %#v", index, policy)
		}
	}
}

func TestReleaseManifestContainsBillingSubscriptionContracts(t *testing.T) {
	manifest, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	job, err := manifest.Jobs().Resolve("billing.subscription.boundary", 1)
	if err != nil || job.OwnerDomain != "billing" ||
		job.EffectClass != jobsv1.EffectDatabaseTransactional {
		t.Fatalf("billing subscription job = %#v/%v", job, err)
	}
	producer, err := manifest.Producers().Resolve("billing.subscription.changed", 1)
	if err != nil {
		t.Fatal(err)
	}
	document, err := os.ReadFile("../../events/v1/billing-subscription-change-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(document))
	if producer.OwnerDomain != "billing" || producer.PayloadSchemaSHA256 != digest {
		t.Fatalf("billing subscription producer = %#v", producer)
	}
}

func assertSameDigest(
	t *testing.T,
	left func() (string, error),
	right func() (string, error),
	name string,
) {
	t.Helper()
	leftDigest, leftErr := left()
	rightDigest, rightErr := right()
	if leftErr != nil || rightErr != nil || leftDigest != rightDigest {
		t.Fatalf("%s manifest drift = %s/%s (%v/%v)",
			name, leftDigest, rightDigest, leftErr, rightErr)
	}
}
