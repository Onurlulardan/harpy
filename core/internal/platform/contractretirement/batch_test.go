package contractretirement

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func testKey(t *testing.T, token string) Key {
	t.Helper()
	key, err := ParseKey("rq.alias." + strings.Repeat(token, 64))
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func testIdentity(t *testing.T) Identity {
	t.Helper()
	identity, err := NewIdentity("restquery", "forms")
	if err != nil {
		t.Fatal(err)
	}
	return identity
}

func TestBatchCanonicalizesAndOwnsKeys(t *testing.T) {
	first, second := testKey(t, "a"), testKey(t, "b")
	input := []Key{second, first}
	batch, err := NewBatch(testIdentity(t), input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = Key{}
	_, values, err := batch.DatabaseValues()
	if err != nil {
		t.Fatal(err)
	}
	firstValue, _ := first.DatabaseValue()
	secondValue, _ := second.DatabaseValue()
	if len(values) != 2 || values[0] != firstValue || values[1] != secondValue {
		t.Fatalf("canonical values = %v", values)
	}
	values[0] = "changed"
	_, owned, _ := batch.DatabaseValues()
	if owned[0] != firstValue {
		t.Fatal("database value mutation escaped batch")
	}
}

func TestBatchRejectsDuplicatesAndBounds(t *testing.T) {
	key := testKey(t, "a")
	if _, err := NewBatch(testIdentity(t), []Key{key, key}); err == nil {
		t.Fatal("duplicate keys were accepted")
	}
	if _, err := NewBatch(testIdentity(t), nil); err == nil {
		t.Fatal("empty batch was accepted")
	}
	tooMany := make([]Key, batchLimit+1)
	for index := range tooMany {
		value := "rq.alias." + fmt.Sprintf("%064x", index)
		tooMany[index], _ = ParseKey(value)
	}
	if _, err := NewBatch(testIdentity(t), tooMany); err == nil {
		t.Fatal("oversized batch was accepted")
	}
}

func TestContractRetirementValuesAreRedacted(t *testing.T) {
	key := testKey(t, "a")
	batch, _ := NewBatch(testIdentity(t), []Key{key})
	decision, _ := SealDecision(batch, nil)
	for _, value := range []any{key, batch, decision, testIdentity(t)} {
		if rendered := fmt.Sprint(value); rendered != redaction {
			t.Fatalf("rendered value = %q", rendered)
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatal("direct JSON marshal succeeded")
		}
	}
}
