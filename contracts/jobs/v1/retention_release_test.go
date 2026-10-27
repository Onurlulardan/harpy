package jobsv1

import (
	"bytes"
	"strings"
	"testing"
)

func TestRetentionReleaseCanonicalDigestIsStable(t *testing.T) {
	value, err := NewRetentionRelease(testRetentionRelease())
	if err != nil {
		t.Fatal(err)
	}
	first, err := value.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	second, err := value.CanonicalJSON()
	if err != nil || !bytes.Equal(first, second) {
		t.Fatal("retention release canonical JSON is not deterministic")
	}
	if !bytes.HasPrefix(first, []byte(`{"authority":`)) {
		t.Fatalf("retention release is not RFC 8785 ordered: %s", first)
	}
	digest, err := value.SHA256()
	if err != nil {
		t.Fatal(err)
	}
	if digest != "5fd2c65bfa1954fad42e5b6c6e3d97a672570318b9b273cf3aaaca32dca8f954" {
		t.Fatalf("retention release digest = %s", digest)
	}
}

func TestRetentionReleaseDecodeIsStrictAndTenantBound(t *testing.T) {
	value, err := NewRetentionRelease(testRetentionRelease())
	if err != nil {
		t.Fatal(err)
	}
	document, err := value.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeRetentionRelease(document); err != nil {
		t.Fatalf("decode canonical release: %v", err)
	}
	unknown := bytes.Replace(document, []byte(`{"authority":`), []byte(`{"extra":true,"authority":`), 1)
	if _, err := DecodeRetentionRelease(unknown); err == nil {
		t.Fatal("retention release accepted an unknown field")
	}
	duplicate := bytes.Replace(document, []byte(`{"authority":`), []byte(`{"job_id":"22222222-2222-4222-8222-222222222222","authority":`), 1)
	if _, err := DecodeRetentionRelease(duplicate); err == nil ||
		!strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("retention release duplicate-field error = %v", err)
	}

	mismatch := testRetentionRelease()
	otherTenant := retryJobID
	mismatch.TenantID = &otherTenant
	if _, err := NewRetentionRelease(mismatch); err == nil {
		t.Fatal("retention release accepted cross-tenant authority")
	}
}

func TestRetentionReleaseRequiresZeroLiveReferencesAndMonotonicRevision(t *testing.T) {
	for name, mutate := range map[string]func(*RetentionRelease){
		"revision":   func(value *RetentionRelease) { value.ReleaseRevision = 0 },
		"dependency": func(value *RetentionRelease) { value.LiveDependencyCount = 1 },
		"retry":      func(value *RetentionRelease) { value.RetryChildCount = 1 },
		"attempt":    func(value *RetentionRelease) { value.UnfinishedAttemptCount = 1 },
		"effect":     func(value *RetentionRelease) { value.LiveEffectReceiptCount = 1 },
	} {
		t.Run(name, func(t *testing.T) {
			value := testRetentionRelease()
			mutate(&value)
			if _, err := NewRetentionRelease(value); err == nil {
				t.Fatal("invalid retention release was accepted")
			}
		})
	}
}

func testRetentionRelease() RetentionRelease {
	tenant := tenantID
	return RetentionRelease{
		JobID: jobID, TenantID: &tenant, Authority: testAuthority(),
		ReleaseRevision: 1, RetainUntil: "2026-10-13T10:00:00Z",
	}
}
