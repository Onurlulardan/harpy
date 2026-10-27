package backupcrypto

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"testing"
)

func TestDecryptBoundSelectsRecoveryKeyFromAuthenticatedHeader(t *testing.T) {
	context := Context{
		EnvironmentID: "018f0000-0000-7000-8000-000000000001",
		RecoverySetID: "018f0000-0000-7000-8000-000000000002",
		ComponentID:   "database", DRKeyID: "dr_kek_g1",
		RecoveryKeyID: "recovery_g1",
	}
	private, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	if _, err = Encrypt(bytes.NewBufferString("payload"), &archive, context,
		SealKeys{DRKEK: bytes.Repeat([]byte{7}, 32),
			RecoveryPublic: private.PublicKey().Bytes(), Random: rand.Reader}); err != nil {
		t.Fatal(err)
	}
	var plaintext bytes.Buffer
	_, actual, err := DecryptBound(&archive, &plaintext, BoundContext{
		EnvironmentID: context.EnvironmentID,
		RecoverySetID: context.RecoverySetID,
		ComponentID:   context.ComponentID,
	}, func(_ string, recoveryKeyID string) (OpenKeys, bool) {
		return OpenKeys{RecoveryPrivate: private.Bytes()},
			recoveryKeyID == context.RecoveryKeyID
	})
	if err != nil || actual != context || plaintext.String() != "payload" {
		t.Fatalf("bound decrypt = %q %#v %v", plaintext.String(), actual, err)
	}
}
