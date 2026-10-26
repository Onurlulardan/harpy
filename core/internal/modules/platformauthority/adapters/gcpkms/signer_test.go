package gcpkms

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"hash/crc32"
	"testing"
)

const fixtureKeyID = "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
	"keyRings/authority/cryptoKeys/proposer/cryptoKeyVersions/7"

type fakeClient struct {
	public     PublicKey
	privateKey ed25519.PrivateKey
	mutate     func(*AsymmetricSignResponse)
	request    AsymmetricSignRequest
}

func (client *fakeClient) GetPublicKey(context.Context, string) (PublicKey, error) {
	return client.public, nil
}

func (client *fakeClient) AsymmetricSign(
	_ context.Context,
	request AsymmetricSignRequest,
) (AsymmetricSignResponse, error) {
	client.request = request
	signature := ed25519.Sign(client.privateKey, request.Data)
	response := AsymmetricSignResponse{
		Name: request.Name, Signature: signature, VerifiedDataCRC32C: true,
		SignatureCRC32C: crc32.Checksum(signature, crc32.MakeTable(crc32.Castagnoli)),
	}
	if client.mutate != nil {
		client.mutate(&response)
	}
	return response, nil
}

func TestSignerUsesVersionedDataAndVerifiesIntegrity(t *testing.T) {
	client, pin := fixtureClient()
	signer, err := NewSigner(context.Background(), fixtureKeyID, pin, client)
	if err != nil {
		t.Fatal(err)
	}
	message := []byte("harpy.protected-authority.proposal.v1\x00{}")
	signature, err := signer.Sign(context.Background(), message)
	if err != nil || !ed25519.Verify(client.public.Key, message, signature) {
		t.Fatalf("KMS signature failed: %v", err)
	}
	resource := "projects/harpy-prod/locations/europe-west1/keyRings/authority/" +
		"cryptoKeys/proposer/cryptoKeyVersions/7"
	if client.request.Name != resource || string(client.request.Data) != string(message) ||
		client.request.DataCRC32C != crc32.Checksum(message, crc32.MakeTable(crc32.Castagnoli)) {
		t.Fatal("AsymmetricSign data/CRC/name contract drifted")
	}
}

func TestSignerRejectsAliasAlgorithmPinAndResponseIntegrity(t *testing.T) {
	client, pin := fixtureClient()
	invalidIDs := []string{
		"gcp-kms://projects/harpy-prod/locations/europe-west1/keyRings/authority/cryptoKeys/proposer",
		"gcp-kms://projects/harpy-prod/locations/europe-west1/keyRings/authority/cryptoKeys/proposer/cryptoKeyVersions/current",
	}
	for _, keyID := range invalidIDs {
		if _, err := NewSigner(context.Background(), keyID, pin, client); err == nil {
			t.Fatal("unversioned or alias KMS key was accepted")
		}
	}
	wrongAlgorithm, _ := fixtureClient()
	wrongAlgorithm.public.Algorithm = "GOOGLE_SYMMETRIC_ENCRYPTION"
	if _, err := NewSigner(context.Background(), fixtureKeyID, pin, wrongAlgorithm); err == nil {
		t.Fatal("wrong KMS algorithm was accepted")
	}
	if _, err := NewSigner(context.Background(), fixtureKeyID, string(make([]byte, 64)), client); err == nil {
		t.Fatal("wrong public key pin was accepted")
	}
	mutations := []func(*AsymmetricSignResponse){
		func(response *AsymmetricSignResponse) { response.Name += "-wrong" },
		func(response *AsymmetricSignResponse) { response.VerifiedDataCRC32C = false },
		func(response *AsymmetricSignResponse) { response.SignatureCRC32C++ },
		func(response *AsymmetricSignResponse) { response.Signature[0] ^= 1 },
	}
	for _, mutate := range mutations {
		candidate, candidatePin := fixtureClient()
		candidate.mutate = mutate
		signer, err := NewSigner(context.Background(), fixtureKeyID, candidatePin, candidate)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := signer.Sign(context.Background(), []byte("message")); err == nil {
			t.Fatal("corrupt KMS response was accepted")
		}
	}
}

func fixtureClient() (*fakeClient, string) {
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = 7
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	digest := sha256.Sum256(publicKey)
	resource := "projects/harpy-prod/locations/europe-west1/keyRings/authority/" +
		"cryptoKeys/proposer/cryptoKeyVersions/7"
	return &fakeClient{
		public: PublicKey{Name: resource, Algorithm: AlgorithmECSignEd25519,
			Purpose: PurposeAsymmetricSign, Key: publicKey}, privateKey: privateKey,
	}, hex.EncodeToString(digest[:])
}
