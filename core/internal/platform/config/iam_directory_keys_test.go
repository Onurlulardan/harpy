package config

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestIAMDirectoryKeysAreCanonicalRedactedDefensiveCopies(t *testing.T) {
	activeID := iamFixtureKID(3)
	oldestID := iamFixtureKID(1)
	newerID := iamFixtureKID(2)
	ring := IAMDirectoryKeyring{iamHMACKeyring: iamHMACKeyring{
		purpose: iamPurposeDirectory,
		active:  iamHMACKey{kid: activeID, material: [32]byte{3}},
		retained: map[string][32]byte{
			newerID: {2}, oldestID: {1},
		},
		valid: true,
	}}
	keys, err := ring.Keys()
	if err != nil || len(keys) != 3 {
		t.Fatalf("keys = %+v, %v", keys, err)
	}
	for index, want := range []string{activeID, oldestID, newerID} {
		id, _ := keys[index].ID()
		if id != want {
			t.Fatalf("key %d id = %q, want %q", index, id, want)
		}
	}
	if active, _ := keys[0].Active(); !active {
		t.Fatal("first key is not active")
	}
	material, _ := keys[0].Material()
	material[0] = 99
	again, _ := ring.Keys()
	againMaterial, _ := again[0].Material()
	if againMaterial[0] != 3 {
		t.Fatal("caller mutated directory keyring material")
	}
	keys[0].Destroy()
	if _, err := keys[0].Material(); err == nil {
		t.Fatal("destroyed directory key remained readable")
	}
	document, err := json.Marshal(keys[0])
	if err != nil || string(document) != `"[REDACTED]"` || fmt.Sprint(keys[0]) != "[REDACTED]" {
		t.Fatalf("directory key leaked: %s %v", document, err)
	}
}

func TestIAMDirectoryKeysFailClosed(t *testing.T) {
	if _, err := (IAMDirectoryKeyring{}).Keys(); err == nil {
		t.Fatal("zero directory keyring was accepted")
	}
	activeID := iamFixtureKID(1)
	ring := IAMDirectoryKeyring{iamHMACKeyring: iamHMACKeyring{
		purpose:  iamPurposeDirectory,
		active:   iamHMACKey{kid: activeID, material: [32]byte{1}},
		retained: map[string][32]byte{activeID: {2}}, valid: true,
	}}
	if _, err := ring.Keys(); err == nil {
		t.Fatal("active key duplicated in retained set")
	}
}
