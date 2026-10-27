package application

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
)

func (service *Coordinator) activeSignerPublicKey(
	ctx context.Context,
	command operationCommand,
) (ed25519.PublicKey, error) {
	binding, err := service.store.LoadBinding(ctx, command.state.Value.RegistryBindingID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	expectedGeneration := command.state.Value.ProposalSignerGeneration
	if command.rotation != nil {
		expectedGeneration = command.rotation.PreviousSignerGeneration
	}
	if binding.BindingID != command.state.Value.RegistryBindingID ||
		binding.EnvironmentID != command.state.Value.EnvironmentID ||
		binding.DeploymentMode != command.state.Value.DeploymentMode ||
		binding.SignerGeneration != expectedGeneration || binding.SignerStatus != "active" ||
		binding.SignerKID != service.signer.KID() ||
		(command.rotation == nil && !optionalStringEqual(binding.RecoveryTrustBundleSHA256,
			command.state.Value.RecoveryTrustBundleSHA256)) {
		return nil, api.ErrConflict
	}
	publicKey, err := service.signer.PublicKey(ctx)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return nil, api.ErrDependency
	}
	digest := sha256.Sum256(publicKey)
	if hex.EncodeToString(digest[:]) != binding.SignerPublicKeySHA256 {
		return nil, api.ErrConflict
	}
	return publicKey, nil
}
