package backupcrypto

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"testing"
)

func TestFCBACKUP1DualWrapRoundTripAndTamper(t *testing.T) {
	private, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	context := Context{EnvironmentID: "018f0000-0000-7000-8000-000000000001",
		RecoverySetID: "018f0000-0000-7000-8000-000000000002",
		ComponentID:   "database", DRKeyID: "dr_kek_g1", RecoveryKeyID: "recovery_g1"}
	plain := bytes.Repeat([]byte("harpy"), ChunkSize/4)
	var archive bytes.Buffer
	result, err := Encrypt(bytes.NewReader(plain), &archive, context, SealKeys{
		DRKEK: bytes.Repeat([]byte{7}, 32), RecoveryPublic: private.PublicKey().Bytes(), Random: rand.Reader})
	if err != nil || result.PlaintextBytes != int64(len(plain)) || result.CiphertextSHA256 == "" {
		t.Fatalf("encrypt result = %#v, err=%v", result, err)
	}
	for _, keys := range []OpenKeys{{DRKEK: bytes.Repeat([]byte{7}, 32)},
		{RecoveryPrivate: private.Bytes()}} {
		var restored bytes.Buffer
		if _, err = Decrypt(bytes.NewReader(archive.Bytes()), &restored, keys); err != nil ||
			!bytes.Equal(restored.Bytes(), plain) {
			t.Fatalf("dual-wrap restore failed: %v", err)
		}
	}
	tampered := append([]byte(nil), archive.Bytes()...)
	tampered[len(tampered)-8] ^= 1
	if _, err = Decrypt(bytes.NewReader(tampered), new(bytes.Buffer),
		OpenKeys{DRKEK: bytes.Repeat([]byte{7}, 32)}); err == nil {
		t.Fatal("tampered archive was accepted")
	}
}

func TestFCBACKUP1RejectsAmbiguousUnwrapSelection(t *testing.T) {
	if _, err := Decrypt(bytes.NewReader([]byte(Magic)), new(bytes.Buffer), OpenKeys{}); err == nil {
		t.Fatal("missing unwrap selection was accepted")
	}
}

func TestFCBACKUP1RejectsMismatchedExpectedContext(t *testing.T) {
	private, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	context := Context{EnvironmentID: "018f0000-0000-7000-8000-000000000001",
		RecoverySetID: "018f0000-0000-7000-8000-000000000002",
		ComponentID:   "database", DRKeyID: "dr_kek_g1", RecoveryKeyID: "recovery_g1"}
	var archive bytes.Buffer
	if _, err = Encrypt(bytes.NewBufferString("database"), &archive, context, SealKeys{
		DRKEK: bytes.Repeat([]byte{7}, 32), RecoveryPublic: private.PublicKey().Bytes(),
		Random: rand.Reader,
	}); err != nil {
		t.Fatal(err)
	}
	wrong := context
	wrong.RecoverySetID = "018f0000-0000-7000-8000-000000000003"
	if _, err = DecryptExpected(bytes.NewReader(archive.Bytes()), new(bytes.Buffer), wrong,
		OpenKeys{DRKEK: bytes.Repeat([]byte{7}, 32)}); err == nil {
		t.Fatal("archive was opened under a different recovery-set context")
	}
}
