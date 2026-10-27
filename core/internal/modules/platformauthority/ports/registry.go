package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

type Registry interface {
	EnsureBinding(context.Context, artifact.VerifiedBindingProof) error
	EnsureProviderBinding(context.Context, ProviderBinding) error
	ExchangeProposal(context.Context, artifact.SignedProposal) ([]byte, error)
	LatestHead(context.Context, HeadKey) (ExternalHead, error)
}

type ProviderBinding struct {
	BindingID             string `json:"binding_id"`
	EnvironmentID         string `json:"environment_id"`
	SignerGeneration      int64  `json:"signer_generation"`
	SignerKID             string `json:"signer_kid"`
	SignerPublicKeySHA256 string `json:"signer_public_key_sha256"`
}

type HeadKey struct {
	RegistryBindingID string
	ScopeKind         artifact.ScopeKind
	ScopeID           string
}

type ExternalHead struct {
	Key              HeadKey
	Generation       int64
	StateSHA256      string
	AckSHA256        string
	RegistrySequence int64
}
