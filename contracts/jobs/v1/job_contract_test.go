package jobsv1

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestCatalogIsDeterministicAndDefensive(t *testing.T) {
	first, _ := NewKindPolicy("zeta.job", 1, "forms", EffectExternal, 8)
	second, _ := NewKindPolicy("alpha.job", 1, "forms", EffectDatabaseTransactional, 4)
	input := []KindPolicy{first, second}
	catalog, err := NewCatalog(input)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := catalog.SHA256()
	input[0].Kind = "mutated.job"
	output := catalog.Policies()
	output[0].Kind = "also.mutated"
	after, _ := catalog.SHA256()
	if before != after || catalog.Policies()[0].Kind != "alpha.job" {
		t.Fatal("catalog was not deterministic and defensively copied")
	}
	encoded, _ := json.Marshal(first)
	unknown := bytes.Replace(encoded, []byte("{"), []byte(`{"unknown":true,`), 1)
	if _, err := DecodeKindPolicy(unknown); err == nil {
		t.Fatal("unknown policy field was accepted")
	}
}

func TestEnqueueUsesPolicyAndCopiesPointers(t *testing.T) {
	catalog := testCatalog(t, EffectExternal)
	input := testEnqueue(EffectExternal)
	value, err := NewEnqueue(input, catalog)
	if err != nil {
		t.Fatal(err)
	}
	*input.TenantID = retryJobID
	*input.EffectIdempotencyKey = "changed"
	if *value.TenantID != tenantID || *value.EffectIdempotencyKey != "form.export/tenant/111" {
		t.Fatal("enqueue retained caller-owned pointers")
	}
	document, _ := json.Marshal(value)
	if _, err := DecodeEnqueue(document, catalog); err != nil {
		t.Fatal(err)
	}
	unknown := bytes.Replace(document, []byte("{"), []byte(`{"unknown":true,`), 1)
	if _, err := DecodeEnqueue(unknown, catalog); err == nil {
		t.Fatal("unknown enqueue field was accepted")
	}
	duplicate := append([]byte(`{"contract_version":"job-enqueue.v1",`), document[1:]...)
	if _, err := DecodeEnqueue(duplicate, catalog); err == nil {
		t.Fatal("duplicate enqueue field was accepted")
	}
	oversized := append(append([]byte(nil), document...), bytes.Repeat([]byte(" "), MaximumDocumentBytes)...)
	if _, err := DecodeEnqueue(oversized, catalog); err == nil {
		t.Fatal("oversized enqueue document was accepted")
	}
	database := testEnqueue(EffectDatabaseTransactional)
	if _, err := NewEnqueue(database, testCatalog(t, EffectDatabaseTransactional)); err != nil {
		t.Fatal(err)
	}
	database.EffectIdempotencyKey = stringPointer("not-allowed")
	if _, err := NewEnqueue(database, testCatalog(t, EffectDatabaseTransactional)); err == nil {
		t.Fatal("transactional job accepted external effect key")
	}
}

func TestEnqueueEnforcesUTF8ByteLimit(t *testing.T) {
	catalog := testCatalog(t, EffectExternal)
	accepted := testEnqueue(EffectExternal)
	accepted.DedupeKey = strings.Repeat("ş", 256)
	if _, err := NewEnqueue(accepted, catalog); err != nil {
		t.Fatalf("512-byte key rejected: %v", err)
	}
	rejected := testEnqueue(EffectExternal)
	rejected.DedupeKey = strings.Repeat("ş", 257)
	if _, err := NewEnqueue(rejected, catalog); err == nil {
		t.Fatal("514-byte key was accepted")
	}
}

func TestLeaseRequiresExactLeaseAndFence(t *testing.T) {
	catalog := testCatalog(t, EffectExternal)
	value, err := NewLease(testLease(), catalog)
	if err != nil {
		t.Fatal(err)
	}
	if value.State != StateRunning || value.HeartbeatSeconds != HeartbeatSeconds {
		t.Fatal("lease constructor did not seal policy values")
	}
	document, _ := json.Marshal(value)
	badState := bytes.Replace(document, []byte(`"running"`), []byte(`"parked"`), 1)
	if _, err := DecodeLease(badState, catalog); err == nil {
		t.Fatal("unknown lease state was accepted")
	}
	bad := testLease()
	bad.AttemptNumber = 2
	if _, err := NewLease(bad, catalog); err == nil {
		t.Fatal("lease accepted attempt and generation mismatch")
	}
	bad = testLease()
	bad.LeaseExpiresAt = "2026-07-15T10:01:01Z"
	if _, err := NewLease(bad, catalog); err == nil {
		t.Fatal("lease accepted non-60-second interval")
	}
}

func TestAttemptEffectBoundaryAndUnknownOutcome(t *testing.T) {
	catalog := testCatalog(t, EffectExternal)
	success := baseAttempt()
	success.ResultState = StateSucceeded
	success.EffectStartedAt = stringPointer("2026-07-15T10:00:01Z")
	success.EffectReceiptID = stringPointer(receiptID)
	if _, err := NewAttemptOutcome(success, catalog); err != nil {
		t.Fatal(err)
	}
	unknown := baseAttempt()
	unknown.ResultState = StateDead
	unknown.EffectStartedAt = stringPointer("2026-07-15T10:00:01Z")
	unknown.ErrorCode = stringPointer(UnknownOutcomeCode)
	if _, err := NewAttemptOutcome(unknown, catalog); err != nil {
		t.Fatal(err)
	}
	unknown.ResultState = StateQueued
	if _, err := NewAttemptOutcome(unknown, catalog); err == nil {
		t.Fatal("unknown provider outcome was accepted as queued")
	}
	known := baseAttempt()
	known.ResultState = StateQueued
	known.EffectStartedAt = stringPointer("2026-07-15T10:00:01Z")
	known.EffectReceiptID = stringPointer(receiptID)
	known.ErrorClass = errorPointer(ErrorTransient)
	known.ErrorCode = stringPointer("PROVIDER_BUSY")
	known.RedactedErrorSHA256 = stringPointer(hexDigest("b"))
	known.RetryRunAt = stringPointer("2026-07-15T10:00:10Z")
	if _, err := NewAttemptOutcome(known, catalog); err != nil {
		t.Fatal(err)
	}
	missingEvidence := known
	missingEvidence.RedactedErrorSHA256 = nil
	if _, err := NewAttemptOutcome(missingEvidence, catalog); err == nil {
		t.Fatal("failed attempt without redacted error evidence was accepted")
	}
	tooLate := known
	tooLate.RetryRunAt = stringPointer("2026-07-15T10:00:11Z")
	if _, err := NewAttemptOutcome(tooLate, catalog); err == nil {
		t.Fatal("retry exceeded full-jitter cap for attempt one")
	}
	final := known
	final.AttemptNumber, final.LeaseGeneration = 8, 8
	if _, err := NewAttemptOutcome(final, catalog); err == nil {
		t.Fatal("final attempt was scheduled for retry")
	}
	if _, err := json.Marshal(baseAttempt()); err == nil {
		t.Fatal("unvalidated attempt outcome was marshaled")
	}
}

func TestOperatorReceiptExactActionMatrix(t *testing.T) {
	adopt := operatorBase(OperatorAdopt)
	adopt.SourceReference = stringPointer("v1/jobs/legacy-42")
	adopt.ResultJobID = stringPointer(retryJobID)
	if _, err := NewOperatorReceipt(adopt); err != nil {
		t.Fatal(err)
	}
	retry := operatorBase(OperatorRetry)
	retry.SourceJobID = stringPointer(jobID)
	retry.ResultJobID = stringPointer(retryJobID)
	retry.SourceEffectIdempotencyKey = stringPointer("form.export/tenant/111")
	retry.ResultEffectIdempotencyKey = stringPointer("form.export/tenant/111")
	validRetry, err := NewOperatorReceipt(retry)
	if err != nil {
		t.Fatal(err)
	}
	unequal := retry
	unequal.ResultEffectIdempotencyKey = stringPointer("different-effect")
	if _, err := NewOperatorReceipt(unequal); err == nil {
		t.Fatal("job retry changed the effect idempotency key")
	}
	databaseRetry := retry
	databaseRetry.SourceEffectIdempotencyKey = nil
	databaseRetry.ResultEffectIdempotencyKey = nil
	if _, err := NewOperatorReceipt(databaseRetry); err != nil {
		t.Fatalf("transactional retry without external effect keys was rejected: %v", err)
	}
	cancel := operatorBase(OperatorCancel)
	cancel.SourceJobID = stringPointer(jobID)
	if _, err := NewOperatorReceipt(cancel); err != nil {
		t.Fatal(err)
	}
	document, _ := json.Marshal(validRetry)
	badVersion := bytes.Replace(document, []byte(OperatorReceiptContractVersion), []byte("job-operator-receipt.v2"), 1)
	if _, err := DecodeOperatorReceipt(badVersion); err == nil {
		t.Fatal("unknown operator receipt version was accepted")
	}
}
