package recoverykeymaterial

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"testing"
)

const materialEnvironmentID = "018f0000-0000-7000-8000-000000000001"

func TestParseAcceptsExactPurposeBoundMaterial(t *testing.T) {
	document := materialFixture(t, ProviderLocalKeyring, "local/recovery-1")
	material, err := Parse(document, materialExpected(
		ProviderLocalKeyring, "local/recovery-1",
	))
	if err != nil {
		t.Fatal(err)
	}
	if material.DRKeyID() != "dr-kek.g1" ||
		material.RecoveryKeyID() != "recovery-x25519.g1" ||
		material.EnvironmentID() != materialEnvironmentID {
		t.Fatalf("parsed material drifted: %#v", material)
	}
	keys := material.SealKeys(bytes.NewReader(make([]byte, 64)))
	if len(keys.DRKEK) != 32 || len(keys.RecoveryPublic) != 32 {
		t.Fatal("resolved sealing bytes are incomplete")
	}
	material.Clear()
}

func TestParseRejectsUnknownAndBindingDrift(t *testing.T) {
	valid := materialFixture(t, ProviderLocalKeyring, "local/recovery-1")
	mutations := [][]byte{
		canonicalMutation(t, valid, func(value map[string]any) { value["unknown"] = true }),
		canonicalMutation(t, valid, func(value map[string]any) {
			value["purpose"] = "restore_component_open"
		}),
		canonicalMutation(t, valid, func(value map[string]any) {
			value["environment_id"] = "018f0000-0000-7000-8000-000000000009"
		}),
		canonicalMutation(t, valid, func(value map[string]any) { value["generation"] = 2 }),
		canonicalMutation(t, valid, func(value map[string]any) {
			value["material_sha256"] = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}),
		canonicalMutation(t, valid, func(value map[string]any) { value["synthetic"] = true }),
		canonicalMutation(t, valid, func(value map[string]any) {
			value["material"].(map[string]any)["dr_kek"].(map[string]any)["key_bytes_base64url"] = ""
		}),
	}
	for index, document := range mutations {
		if _, err := Parse(document, materialExpected(
			ProviderLocalKeyring, "local/recovery-1",
		)); err == nil {
			t.Fatalf("invalid material mutation %d was accepted", index)
		}
	}
}

func TestParseRejectsWrongProviderAndNonCanonicalBytes(t *testing.T) {
	document := materialFixture(t, ProviderLocalKeyring, "local/recovery-1")
	if _, err := Parse(document, materialExpected(
		ProviderLocalKeyring, "binding/recovery-1",
	)); err == nil {
		t.Fatal("wrong provider binding was accepted")
	}
	if _, err := Parse(append([]byte(" "), document...), materialExpected(
		ProviderLocalKeyring, "local/recovery-1",
	)); err == nil {
		t.Fatal("non-canonical material was accepted")
	}
}

func materialExpected(kind, handle string) Expected {
	return Expected{Purpose: PurposeComponentSealing, ProviderKind: kind,
		ProviderHandle: handle, EnvironmentID: materialEnvironmentID, Generation: 1}
}

func materialFixture(t *testing.T, kind, handle string) []byte {
	t.Helper()
	body := materialBody{
		DRKEK: keyWire{KeyID: "dr-kek.g1", Algorithm: "AES-256",
			Bytes: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))},
		RecoveryPublicKey: keyWire{KeyID: "recovery-x25519.g1", Algorithm: "X25519",
			Bytes: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{9}, 32))},
	}
	bodyBytes, _ := canonicalValue(body)
	digest := sha256.Sum256(bodyBytes)
	wire := materialWire{Schema: Schema, Purpose: PurposeComponentSealing,
		ProviderKind: kind, ProviderHandle: handle, EnvironmentID: materialEnvironmentID,
		Generation: 1, Material: body, MaterialSHA256: hex.EncodeToString(digest[:])}
	return canonicalFixture(t, wire)
}

func canonicalMutation(t *testing.T, document []byte, mutate func(map[string]any)) []byte {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(document, &value); err != nil {
		t.Fatal(err)
	}
	mutate(value)
	return canonicalFixture(t, value)
}

func canonicalFixture(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var normalized any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err = decoder.Decode(&normalized); err != nil {
		t.Fatal(err)
	}
	result, err := json.Marshal(normalized)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
