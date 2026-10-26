package gcpkms

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"hash/crc32"

	"cloud.google.com/go/auth/credentials"
	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type googleKMS interface {
	GetPublicKey(context.Context, *kmspb.GetPublicKeyRequest,
		...gax.CallOption) (*kmspb.PublicKey, error)
	AsymmetricSign(context.Context, *kmspb.AsymmetricSignRequest,
		...gax.CallOption) (*kmspb.AsymmetricSignResponse, error)
	Close() error
}

type GoogleClient struct{ client googleKMS }

func OpenGoogleClient(
	ctx context.Context,
	workloadIdentityPath string,
) (*GoogleClient, error) {
	if ctx == nil || ctx.Err() != nil {
		return nil, fmt.Errorf("SaaS authority KMS context is invalid")
	}
	options := &credentials.DetectOptions{
		Scopes:          []string{"https://www.googleapis.com/auth/cloudkms"},
		CredentialsFile: workloadIdentityPath,
	}
	identity, err := credentials.DetectDefault(options)
	if err != nil || identity == nil ||
		!workloadCredentialDocumentAllowed(identity.JSON()) {
		return nil, fmt.Errorf("SaaS authority workload identity is invalid")
	}
	client, err := kms.NewKeyManagementClient(ctx, option.WithAuthCredentials(identity))
	if err != nil {
		return nil, fmt.Errorf("open SaaS authority KMS client: %w", err)
	}
	return &GoogleClient{client: client}, nil
}

func (client *GoogleClient) Close() error {
	if client == nil || client.client == nil {
		return nil
	}
	return client.client.Close()
}

func (client *GoogleClient) GetPublicKey(
	ctx context.Context,
	name string,
) (PublicKey, error) {
	if client == nil || client.client == nil || name == "" {
		return PublicKey{}, fmt.Errorf("SaaS authority KMS client is unavailable")
	}
	response, err := client.client.GetPublicKey(ctx,
		&kmspb.GetPublicKeyRequest{Name: name})
	if err != nil || response == nil || response.Name != name ||
		response.Algorithm != kmspb.CryptoKeyVersion_EC_SIGN_ED25519 ||
		!validCRC32C([]byte(response.Pem), response.PemCrc32C) {
		return PublicKey{}, fmt.Errorf("read SaaS authority KMS public key: %w", err)
	}
	block, rest := pem.Decode([]byte(response.Pem))
	if block == nil || len(rest) != 0 || block.Type != "PUBLIC KEY" {
		return PublicKey{}, fmt.Errorf("SaaS authority KMS public key is invalid")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	key, valid := parsed.(ed25519.PublicKey)
	if err != nil || !valid || len(key) != ed25519.PublicKeySize {
		return PublicKey{}, fmt.Errorf("SaaS authority KMS public key is invalid")
	}
	return PublicKey{Name: name, Algorithm: AlgorithmECSignEd25519,
		Purpose: PurposeAsymmetricSign, Key: append(ed25519.PublicKey(nil), key...)}, nil
}

func (client *GoogleClient) AsymmetricSign(
	ctx context.Context,
	request AsymmetricSignRequest,
) (AsymmetricSignResponse, error) {
	if client == nil || client.client == nil || request.Name == "" ||
		len(request.Data) == 0 {
		return AsymmetricSignResponse{}, fmt.Errorf("SaaS authority KMS request is invalid")
	}
	response, err := client.client.AsymmetricSign(ctx, &kmspb.AsymmetricSignRequest{
		Name: request.Name, Data: append([]byte(nil), request.Data...),
		DataCrc32C: wrapperspb.Int64(int64(request.DataCRC32C)),
	})
	if err != nil || response == nil || response.SignatureCrc32C == nil {
		return AsymmetricSignResponse{}, fmt.Errorf("sign SaaS authority proposal: %w", err)
	}
	return AsymmetricSignResponse{Name: response.Name,
		Signature:          append([]byte(nil), response.Signature...),
		VerifiedDataCRC32C: response.VerifiedDataCrc32C,
		SignatureCRC32C:    uint32(response.SignatureCrc32C.Value)}, nil
}

func validCRC32C(value []byte, expected *wrapperspb.Int64Value) bool {
	if expected == nil || expected.Value < 0 || expected.Value > int64(^uint32(0)) {
		return false
	}
	return crc32.Checksum(value, crc32.MakeTable(crc32.Castagnoli)) ==
		uint32(expected.Value)
}

var _ Client = (*GoogleClient)(nil)
