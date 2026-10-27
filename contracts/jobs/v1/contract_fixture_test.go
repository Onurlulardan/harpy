package jobsv1

import "testing"

const (
	tenantID    = "11111111-1111-4111-8111-111111111111"
	jobID       = "22222222-2222-4222-8222-222222222222"
	retryJobID  = "33333333-3333-4333-8333-333333333333"
	attemptID   = "44444444-4444-4444-8444-444444444444"
	workerID    = "55555555-5555-4555-8555-555555555555"
	receiptID   = "66666666-6666-4666-8666-666666666666"
	actorID     = "77777777-7777-4777-8777-777777777777"
	causationID = "88888888-8888-4888-8888-888888888888"
	timestamp   = "2026-07-15T10:00:00Z"
)

func testAuthority() Authority {
	return Authority{
		ScopeKind: ScopeTenant, ScopeID: tenantID, Domain: "forms", CutoverGeneration: 3,
		WriterOrigin: WriterV2, GroupID: "billing_forms", GroupGeneration: 4,
	}
}

func testCatalog(t *testing.T, effect EffectClass) Catalog {
	t.Helper()
	policy, err := NewKindPolicy("form.export", 1, "forms", effect, DefaultMaximumAttempts)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCatalog([]KindPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func testEnqueue(effect EffectClass) Enqueue {
	tenant := tenantID
	effectKey := "form.export/tenant/111"
	value := Enqueue{
		JobID: jobID, Kind: "form.export", PayloadRevision: 1, TenantID: &tenant,
		Authority: testAuthority(), PayloadReference: "payload/jobs/222", PayloadSHA256: hexDigest("a"),
		RunAt: timestamp, Priority: 10, DedupeKey: "form.export/111", MaximumAttempts: 8,
	}
	if effect == EffectExternal {
		value.EffectIdempotencyKey = &effectKey
	}
	return value
}

func testLease() Lease {
	tenant := tenantID
	return Lease{
		JobID: jobID, Kind: "form.export", PayloadRevision: 1, TenantID: &tenant,
		Authority: testAuthority(), LeaseOwner: workerID, LeaseGeneration: 1, AttemptNumber: 1,
		LeasedAt: timestamp, LeaseExpiresAt: "2026-07-15T10:01:00Z",
	}
}

func baseAttempt() AttemptOutcome {
	return AttemptOutcome{
		AttemptID: attemptID, JobID: jobID, Kind: "form.export", PayloadRevision: 1,
		AttemptNumber: 1, LeaseOwner: workerID, LeaseGeneration: 1,
		StartedAt: timestamp, CompletedAt: "2026-07-15T10:00:05Z",
	}
}

func operatorBase(action OperatorAction) OperatorReceipt {
	tenant := tenantID
	return OperatorReceipt{
		ReceiptID: receiptID, Action: action, TenantID: &tenant, Authority: testAuthority(),
		Reason: "operator verified durable evidence", TicketReference: "OPS-42",
		ActorID: actorID, CausationID: causationID, OccurredAt: timestamp,
	}
}

func stringPointer(value string) *string { return &value }

func errorPointer(value ErrorClass) *ErrorClass { return &value }

func hexDigest(character string) string {
	value := ""
	for len(value) < 64 {
		value += character
	}
	return value
}
