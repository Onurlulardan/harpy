package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *BindingService) BindOnPremise(
	ctx context.Context,
	request api.OnPremiseBindingRequest,
) (api.BindingReceipt, error) {
	if service.registryKind == artifact.RegistrySaaS {
		return api.BindingReceipt{}, api.ErrBindingConflict
	}
	publicKey, err := service.signer.PublicKey(ctx)
	if err != nil {
		return api.BindingReceipt{}, api.ErrDependency
	}
	proof, err := artifact.VerifyBindingProof(request.ProofBytes, publicKey)
	if err != nil || proof.Value.LicenseBindingID != request.BindingID ||
		proof.Value.EnvironmentID != request.EnvironmentID ||
		proof.Value.FinalizationReceiptSHA256 != request.FinalizationReceiptSHA256 ||
		proof.Value.RecoveryTrustBundleSHA256 != request.RecoveryTrustBundleSHA256 ||
		proof.Value.ControlKID != service.signer.KID() {
		return api.BindingReceipt{}, api.ErrBindingConflict
	}
	publicDigest := sha256.Sum256(publicKey)
	licenseID := request.BindingID
	finalization := request.FinalizationReceiptSHA256
	recovery := request.RecoveryTrustBundleSHA256
	record := ports.BindingRecord{
		BindingID: request.BindingID, EnvironmentID: request.EnvironmentID,
		DeploymentMode: artifact.ModeOnPremise, RegistryKind: service.registryKind,
		LicenseBindingID: &licenseID, FinalizationReceiptSHA256: &finalization,
		RecoveryTrustBundleSHA256: &recovery, SignerGeneration: 1,
		SignerKID:             service.signer.KID(),
		SignerPublicKeySHA256: hex.EncodeToString(publicDigest[:]),
	}
	if err := service.registry.EnsureBinding(ctx, proof); err != nil {
		return api.BindingReceipt{}, api.ErrDependency
	}
	if err := service.store.SaveBinding(ctx, record, &proof); err != nil {
		return api.BindingReceipt{}, mapStoreError(err)
	}
	return bindingReceipt(record), nil
}
