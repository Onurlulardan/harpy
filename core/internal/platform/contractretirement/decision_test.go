package contractretirement

import "testing"

func TestDecisionSealsOnlySubmittedKeys(t *testing.T) {
	first, second := testKey(t, "a"), testKey(t, "b")
	batch, err := NewBatch(testIdentity(t), []Key{first, second})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := SealDecision(batch, []Key{second})
	if err != nil {
		t.Fatal(err)
	}
	retired, err := decision.Retires(second)
	if err != nil || !retired {
		t.Fatalf("retired decision = %t err=%v", retired, err)
	}
	retired, err = decision.Retires(first)
	if err != nil || retired {
		t.Fatalf("active decision = %t err=%v", retired, err)
	}
	outside := testKey(t, "c")
	if _, err := decision.Retires(outside); err == nil {
		t.Fatal("decision accepted a key outside submitted batch")
	}
	if _, err := SealDecision(batch, []Key{outside}); err == nil {
		t.Fatal("decision sealed a key outside submitted batch")
	}
}

func TestDecisionRejectsDuplicateRetiredRows(t *testing.T) {
	key := testKey(t, "a")
	batch, _ := NewBatch(testIdentity(t), []Key{key})
	if _, err := SealDecision(batch, []Key{key, key}); err == nil {
		t.Fatal("duplicate retired rows were accepted")
	}
}
