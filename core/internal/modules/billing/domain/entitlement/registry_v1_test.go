package entitlement

import "testing"

func TestRegistryV1ExactShape(t *testing.T) {
	want := []struct {
		key     Key
		kind    ValueKind
		meter   MeterKind
		reset   ResetPolicy
		enforce Enforcement
		unit    Unit
		owner   string
	}{
		{"forms.submissions", Integer, Counter, CalendarMonthUTC, Hard, UnitCount, "forms"},
		{"iam.active_users", Integer, Gauge, Never, Hard, UnitCount, "iam"},
		{"forms.active_forms", Integer, Gauge, Never, Hard, UnitCount, "forms"},
		{"documents.storage_bytes", Integer, Gauge, Never, Hard, UnitByte, "documents"},
		{"ai.chat", Boolean, NoMeter, Never, NoEnforcement, UnitBoolean, "ai"},
		{"workflow.enabled", Boolean, NoMeter, Never, NoEnforcement, UnitBoolean, "workflow"},
		{"api.access", Boolean, NoMeter, Never, NoEnforcement, UnitBoolean, "platform"},
	}
	entries := RegistryV1()
	if len(entries) != len(want) {
		t.Fatalf("registry length = %d", len(entries))
	}
	seen := map[Key]struct{}{}
	for index, expected := range want {
		entry := entries[index]
		definition := entry.Definition
		if definition.Key() != expected.key || definition.ValueKind() != expected.kind ||
			definition.MeterKind() != expected.meter ||
			definition.ResetPolicy() != expected.reset ||
			definition.Enforcement() != expected.enforce ||
			definition.AllowsUnlimited() != (expected.kind == Integer) ||
			entry.Unit != expected.unit || entry.Owner != expected.owner {
			t.Fatalf("entry %d mismatch: %#v", index, entry)
		}
		if _, exists := seen[definition.Key()]; exists {
			t.Fatalf("duplicate key %q", definition.Key())
		}
		seen[definition.Key()] = struct{}{}
	}
	if len(RegistryV1ByKey()) != len(want) {
		t.Fatal("registry lookup lost entries")
	}
}

func TestKeyMatchesStorageShape(t *testing.T) {
	valid := []string{"a", "ai.chat", "module.feature_2", "a.b"}
	invalid := []string{"A.B", "a..", "-a", "module.feature-2", "a/feature"}
	for _, value := range valid {
		if _, err := NewKey(value); err != nil {
			t.Fatalf("valid key %q: %v", value, err)
		}
	}
	for _, value := range invalid {
		if _, err := NewKey(value); err == nil {
			t.Fatalf("invalid key %q accepted", value)
		}
	}
}
