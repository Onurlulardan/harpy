package gcpkms

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/crc32"

	kmsv1 "github.com/Onurlulardan/harpy/contracts/kms/v1"
)

type Signer struct {
	kid       string
	resource  string
	publicKey ed25519.PublicKey
	client    Client
}

func NewSigner(
	ctx context.Context,
	keyID string,
	expectedPublicKeySHA256 string,
	client Client,
) (*Signer, error) {
	version, err := kmsv1.ParseGoogleCloudKeyVersion(keyID)
	if err != nil || client == nil || len(expectedPublicKeySHA256) != 64 {
		return nil, fmt.Errorf("SaaS authority KMS configuration is invalid")
	}
	resource, _ := version.Resource()
	metadata, err := client.GetPublicKey(ctx, resource)
	if err != nil || metadata.Name != resource ||
		metadata.Algorithm != AlgorithmECSignEd25519 ||
		metadata.Purpose != PurposeAsymmetricSign || len(metadata.Key) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("SaaS authority KMS metadata is invalid")
	}
	digest := sha256.Sum256(metadata.Key)
	if hex.EncodeToString(digest[:]) != expectedPublicKeySHA256 {
		return nil, fmt.Errorf("SaaS authority KMS public key pin mismatches")
	}
	return &Signer{
		kid: keyID, resource: resource,
		publicKey: append(ed25519.PublicKey(nil), metadata.Key...), client: client,
	}, nil
}

func (signer *Signer) KID() string { return signer.kid }

func (signer *Signer) PublicKey(context.Context) (ed25519.PublicKey, error) {
	return append(ed25519.PublicKey(nil), signer.publicKey...), nil
}

func (signer *Signer) Sign(ctx context.Context, message []byte) ([]byte, error) {
	if len(message) == 0 || len(message) > 256*1024 {
		return nil, fmt.Errorf("SaaS authority KMS signing input is invalid")
	}
	table := crc32.MakeTable(crc32.Castagnoli)
	response, err := signer.client.AsymmetricSign(ctx, AsymmetricSignRequest{
		Name: signer.resource, Data: append([]byte(nil), message...),
		DataCRC32C: crc32.Checksum(message, table),
	})
	if err != nil || response.Name != signer.resource || !response.VerifiedDataCRC32C ||
		len(response.Signature) != ed25519.SignatureSize ||
		crc32.Checksum(response.Signature, table) != response.SignatureCRC32C ||
		!ed25519.Verify(signer.publicKey, message, response.Signature) {
		return nil, fmt.Errorf("SaaS authority KMS integrity verification failed")
	}
	return append([]byte(nil), response.Signature...), nil
}
