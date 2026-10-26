package gcpkms

import (
	"context"
	"crypto/ed25519"
)

const (
	AlgorithmECSignEd25519 = "EC_SIGN_ED25519"
	PurposeAsymmetricSign  = "ASYMMETRIC_SIGN"
	RequiredIAMPermission  = "cloudkms.cryptoKeyVersions.useToSign"
)

type PublicKey struct {
	Name      string
	Algorithm string
	Purpose   string
	Key       ed25519.PublicKey
}

type AsymmetricSignRequest struct {
	Name       string
	Data       []byte
	DataCRC32C uint32
}

type AsymmetricSignResponse struct {
	Name               string
	Signature          []byte
	VerifiedDataCRC32C bool
	SignatureCRC32C    uint32
}

type Client interface {
	GetPublicKey(context.Context, string) (PublicKey, error)
	AsymmetricSign(context.Context, AsymmetricSignRequest) (AsymmetricSignResponse, error)
}
