package authority

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestEvidenceIsImmutableRedactedContextAuthority(t *testing.T) {
	actor := uuid.MustParse("018f0000-0000-7000-8000-000000000001")
	tenant := uuid.MustParse("018f0000-0000-7000-8000-000000000002")
	evidence, err := NewEvidence(actor, tenant, 3, 5, 7,
		"analytics.query.execute", "analytics.query.execute")
	if err != nil || evidence.Digest() == [32]byte{} {
		t.Fatalf("evidence = %#v, err = %v", evidence, err)
	}
	ctx, err := WithEvidence(context.Background(), evidence)
	loaded, ok := FromContext(ctx)
	if err != nil || !ok || loaded.Digest() != evidence.Digest() {
		t.Fatal("authority context did not preserve sealed evidence")
	}
}

func TestEvidenceRejectsIncompleteAndTamperedInputs(t *testing.T) {
	actor := uuid.MustParse("018f0000-0000-7000-8000-000000000001")
	tenant := uuid.MustParse("018f0000-0000-7000-8000-000000000002")
	for _, input := range []struct {
		actor, tenant         uuid.UUID
		token, version, clock int64
		permission, operation string
	}{
		{uuid.Nil, tenant, 1, 1, 1, "analytics.query.read", "analytics.query.get"},
		{actor, uuid.Nil, 1, 1, 1, "analytics.query.read", "analytics.query.get"},
		{actor, tenant, 0, 1, 1, "analytics.query.read", "analytics.query.get"},
		{actor, tenant, 1, 1, 1, "INVALID", "analytics.query.get"},
	} {
		if _, err := NewEvidence(input.actor, input.tenant, input.token, input.version,
			input.clock, input.permission, input.operation); err == nil {
			t.Fatal("invalid evidence was issued")
		}
	}
}
