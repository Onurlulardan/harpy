package cryptography

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
)

func TestWebAuthnVerifiesSignedAssertionAndUserHandle(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	challenge := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
	rpID, origin := "app.example.com", "https://app.example.com"
	credentialID := []byte("credential-id-v1")
	userID := uuid.MustParse("51000000-0000-4000-8000-000000000001")
	publicKey := coseES256PublicKey(t, &privateKey.PublicKey)
	wire := signedAssertionWire(
		t, privateKey, challenge, rpID, origin, credentialID, userID[:], 7,
	)
	assertion, err := NewWebAuthnAssertion(wire)
	clear(wire)
	if err != nil {
		t.Fatal(err)
	}
	defer assertion.Destroy()
	expectation := authentication.WebAuthnExpectation{
		Challenge: []byte(challenge), CredentialID: credentialID,
		PublicKeyCOSE: publicKey, RPID: rpID, Origin: origin,
		UserHandle: userID[:], UserVerificationRequired: true,
	}
	result, err := NewWebAuthn().Verify(assertion, expectation)
	if err != nil || result.SignCount != 7 || !result.UserVerified {
		t.Fatalf("WebAuthn result=%+v err=%v", result, err)
	}
	otherUserID := uuid.MustParse(
		"51000000-0000-4000-8000-000000000002",
	)
	expectation.UserHandle = otherUserID[:]
	_, err = NewWebAuthn().Verify(assertion, expectation)
	if !errors.Is(err, authentication.ErrInvalidCredentials) {
		t.Fatalf("substituted user handle error=%v", err)
	}
}

func coseES256PublicKey(t *testing.T, key *ecdsa.PublicKey) []byte {
	t.Helper()
	x := key.X.FillBytes(make([]byte, 32))
	y := key.Y.FillBytes(make([]byte, 32))
	encoded, err := cbor.Marshal(map[int]any{
		1: 2, 3: -7, -1: 1, -2: x, -3: y,
	})
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func signedAssertionWire(
	t *testing.T, key *ecdsa.PrivateKey, challenge, rpID, origin string,
	credentialID, userHandle []byte, counter uint32,
) []byte {
	t.Helper()
	clientData, err := json.Marshal(map[string]any{
		"type": "webauthn.get", "challenge": challenge,
		"origin": origin, "crossOrigin": false,
	})
	if err != nil {
		t.Fatal(err)
	}
	authenticatorData := make([]byte, 37)
	rpHash := sha256.Sum256([]byte(rpID))
	copy(authenticatorData, rpHash[:])
	authenticatorData[32] = 0x05
	binary.BigEndian.PutUint32(authenticatorData[33:], counter)
	clientHash := sha256.Sum256(clientData)
	signed := append(append([]byte(nil), authenticatorData...), clientHash[:]...)
	digest := sha256.Sum256(signed)
	signature, err := ecdsa.SignASN1(rand.Reader, key, digest[:])
	clear(signed)
	if err != nil {
		t.Fatal(err)
	}
	encode := base64.RawURLEncoding.EncodeToString
	wire, err := json.Marshal(map[string]any{
		"id": encode(credentialID), "rawId": encode(credentialID),
		"type": "public-key", "clientExtensionResults": map[string]any{},
		"response": map[string]any{
			"authenticatorData": encode(authenticatorData),
			"clientDataJSON":    encode(clientData), "signature": encode(signature),
			"userHandle": encode(userHandle),
		},
	})
	clear(clientData)
	clear(authenticatorData)
	clear(signature)
	if err != nil {
		t.Fatal(err)
	}
	return wire
}
