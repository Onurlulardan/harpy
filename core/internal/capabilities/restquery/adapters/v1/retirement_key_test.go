package v1query

import (
	"encoding/json"
	"fmt"
	"testing"
)

const retirementFixtureManifest = "a1111111-1111-4111-8111-111111111111"

func TestRetirementKeyDomainSeparatedGolden(t *testing.T) {
	key, err := NewRetirementKey(retirementFixtureManifest, "query.search.v1")
	if err != nil {
		t.Fatal(err)
	}
	value, err := key.Value()
	if err != nil {
		t.Fatal(err)
	}
	want := "rq.alias.03cef2893997f4c67dcfa7290c872eb4" +
		"7119c250b09e166cc60a7e1dbefdaca6"
	if value != want {
		t.Fatalf("retirement key=%q", value)
	}
	otherAlias, _ := NewRetirementKey(retirementFixtureManifest, "sort.status.v1")
	otherManifest, _ := NewRetirementKey(
		"a2222222-2222-4222-8222-222222222222", "query.search.v1",
	)
	aliasValue, _ := otherAlias.Value()
	manifestValue, _ := otherManifest.Value()
	if value == aliasValue || value == manifestValue || aliasValue == manifestValue {
		t.Fatal("distinct manifest-alias pairs shared retirement identity")
	}
}

func TestRetirementKeyRejectsInvalidAndTamperedState(t *testing.T) {
	for _, input := range []struct{ manifest, alias string }{
		{"", "query.search.v1"},
		{retirementFixtureManifest, ""},
		{retirementFixtureManifest, "BAD"},
	} {
		if _, err := NewRetirementKey(input.manifest, input.alias); err == nil {
			t.Fatalf("invalid input accepted: %#v", input)
		}
	}
	if (RetirementKey{}).Validate() == nil {
		t.Fatal("zero retirement key validated")
	}
	key, _ := NewRetirementKey(retirementFixtureManifest, "query.search.v1")
	key.value = "rq.alias." + key.value[len("rq.alias.")+1:]
	if key.Validate() == nil {
		t.Fatal("tampered retirement key validated")
	}
}

func TestRetirementKeyRedactsAndRejectsDirectJSON(t *testing.T) {
	key, _ := NewRetirementKey(retirementFixtureManifest, "query.search.v1")
	if fmt.Sprint(key) != v1Redaction || fmt.Sprintf("%#v", key) != v1Redaction {
		t.Fatal("retirement key formatting leaked state")
	}
	if _, err := json.Marshal(key); err == nil {
		t.Fatal("retirement key marshaled directly")
	}
}
