package invitation

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type credentialKeyringFixture struct {
	id  string
	key [32]byte
}

func (fixture credentialKeyringFixture) Active() (string, [32]byte, error) {
	return fixture.id, fixture.key, nil
}

func (fixture credentialKeyringFixture) HMACKey(id string) ([32]byte, error) {
	if id != fixture.id {
		return [32]byte{}, errors.New("unknown key")
	}
	return fixture.key, nil
}

func TestInvitationCredentialRoundTripIsPurposeAndTenantBound(t *testing.T) {
	tenantID := uuid.New()
	keys := credentialKeyringFixture{id: fixtureTokenKeyID(), key: [32]byte{9}}
	generator := CredentialGenerator{randomness: bytes.NewReader(bytes.Repeat([]byte{7}, 48))}
	issued, evidence, err := generator.GenerateCredential(tenantID, keys)
	if err != nil {
		t.Fatal(err)
	}
	var wire []byte
	if err := issued.RevealOnce(func(value []byte) error {
		wire = append([]byte(nil), value...)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	presented, err := ParseCredential(wire)
	clear(wire)
	if err != nil {
		t.Fatal(err)
	}
	defer presented.Destroy()
	proof, err := presented.Proof(
		tenantID, uuid.New(), 1, evidence.HashKeyID, evidence.SecretHMAC, keys,
	)
	if err != nil || proof.BindingSHA256 != evidence.BindingSHA256 ||
		proof.Selector != evidence.Selector {
		t.Fatalf("proof = %#v, %v", proof, err)
	}
	if _, err := presented.Proof(
		uuid.New(), uuid.New(), 1, evidence.HashKeyID, evidence.SecretHMAC, keys,
	); !errors.Is(err, ErrCredentialInvalid) {
		t.Fatalf("cross-tenant proof result = %v", err)
	}
}

func TestInvitationCredentialRejectsTamperAndDestroysOnce(t *testing.T) {
	keys := credentialKeyringFixture{id: fixtureTokenKeyID(), key: [32]byte{1}}
	generator := CredentialGenerator{randomness: bytes.NewReader(bytes.Repeat([]byte{5}, 48))}
	issued, _, err := generator.GenerateCredential(uuid.New(), keys)
	if err != nil {
		t.Fatal(err)
	}
	var callback []byte
	var copied []byte
	if err := issued.RevealOnce(func(value []byte) error {
		callback = value
		copied = append([]byte(nil), value...)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(callback, make([]byte, credentialWireBytes)) {
		t.Fatal("temporary credential wire was not wiped")
	}
	if err := issued.RevealOnce(func([]byte) error { return nil }); !errors.Is(err, ErrCredentialDestroyed) {
		t.Fatalf("second reveal result = %v", err)
	}
	copied[len(copied)-1] ^= 1
	if _, err := ParseCredential(copied); !errors.Is(err, ErrCredentialInvalid) {
		t.Fatalf("tampered credential result = %v", err)
	}
}
