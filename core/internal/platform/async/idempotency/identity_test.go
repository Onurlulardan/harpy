package idempotency

import (
	"testing"

	"github.com/google/uuid"
)

func TestIdentityAdvisoryKeyIsDeterministicAndFullyScoped(t *testing.T) {
	t.Parallel()
	tenant := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	principal := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	identity, err := NewIdentity(tenant, principal, "forms.record.create", "request-1")
	if err != nil {
		t.Fatal(err)
	}
	again, _ := NewIdentity(tenant, principal, "forms.record.create", "request-1")
	if identity.AdvisoryKey() != again.AdvisoryKey() {
		t.Fatal("same identity produced different advisory keys")
	}
	variants := []Identity{}
	for _, input := range []struct {
		tenant, principal uuid.UUID
		operation, key    string
	}{
		{uuid.New(), principal, "forms.record.create", "request-1"},
		{tenant, uuid.New(), "forms.record.create", "request-1"},
		{tenant, principal, "forms.record.update", "request-1"},
		{tenant, principal, "forms.record.create", "request-2"},
	} {
		value, valueErr := NewIdentity(input.tenant, input.principal, input.operation, input.key)
		if valueErr != nil {
			t.Fatal(valueErr)
		}
		variants = append(variants, value)
	}
	for _, variant := range variants {
		if identity.AdvisoryKey() == variant.AdvisoryKey() {
			t.Fatal("scoped identity variant reused advisory key")
		}
	}
}

func TestIdentityRejectsInvalidAndOversizedValues(t *testing.T) {
	t.Parallel()
	valid := uuid.New()
	for _, test := range []struct {
		tenant, principal uuid.UUID
		operation, key    string
	}{
		{uuid.Nil, valid, "forms.create", "key"},
		{valid, uuid.Nil, "forms.create", "key"},
		{valid, valid, "Forms.create", "key"},
		{valid, valid, "forms.create", ""},
		{valid, valid, "forms.create", string(make([]byte, 513))},
	} {
		if _, err := NewIdentity(test.tenant, test.principal, test.operation, test.key); err == nil {
			t.Fatal("invalid identity was accepted")
		}
	}
}
