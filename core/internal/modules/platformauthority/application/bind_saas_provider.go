package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *BindingService) BindSaaSProvider(
	ctx context.Context,
	request api.SaaSProviderBindingRequest,
) (api.BindingReceipt, error) {
	if service.registryKind != artifact.RegistrySaaS || request.SignerGeneration < 1 {
		return api.BindingReceipt{}, api.ErrBindingConflict
	}
	publicKey, err := service.signer.PublicKey(ctx)
	if err != nil {
		return api.BindingReceipt{}, api.ErrDependency
	}
	publicDigest := sha256.Sum256(publicKey)
	record := ports.BindingRecord{
		BindingID: request.BindingID, EnvironmentID: request.EnvironmentID,
		DeploymentMode: artifact.ModeSaaS, RegistryKind: artifact.RegistrySaaS,
		SignerGeneration: request.SignerGeneration, SignerKID: service.signer.KID(),
		SignerPublicKeySHA256: hex.EncodeToString(publicDigest[:]),
	}
	providerBinding := ports.ProviderBinding{
		BindingID: record.BindingID, EnvironmentID: record.EnvironmentID,
		SignerGeneration: record.SignerGeneration, SignerKID: record.SignerKID,
		SignerPublicKeySHA256: record.SignerPublicKeySHA256,
	}
	if err := service.registry.EnsureProviderBinding(ctx, providerBinding); err != nil {
		return api.BindingReceipt{}, api.ErrDependency
	}
	if err := service.store.SaveBinding(ctx, record, nil); err != nil {
		return api.BindingReceipt{}, mapStoreError(err)
	}
	return bindingReceipt(record), nil
}
