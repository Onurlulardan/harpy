package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *memoryStore) ApplySignerRotation(
	_ context.Context,
	value ports.SignerRotationRecord,
) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	binding, found := store.bindings[value.BindingID]
	if !found {
		return ports.ErrNotFound
	}
	if binding.SignerGeneration == value.NewSignerGeneration &&
		binding.SignerKID == value.NewSignerKID &&
		binding.SignerPublicKeySHA256 == value.NewSignerPublicKeySHA256 &&
		optionalStringEqual(binding.RecoveryTrustBundleSHA256,
			value.NewRecoveryTrustBundleSHA256) &&
		binding.SignerStatus == "active" {
		return nil
	}
	if store.failActivationOnce {
		store.failActivationOnce = false
		return context.DeadlineExceeded
	}
	operation, found := store.operations[value.OperationID]
	if !found || operation.Status != ports.StatusCompleted ||
		operation.Purpose != "proposal_signer_rotation" ||
		binding.SignerGeneration != value.ExpectedSignerGeneration ||
		binding.SignerKID != value.ExpectedSignerKID ||
		binding.SignerStatus != "active" {
		return ports.ErrConflict
	}
	binding.SignerGeneration = value.NewSignerGeneration
	binding.SignerKID = value.NewSignerKID
	binding.SignerPublicKeySHA256 = value.NewSignerPublicKeySHA256
	binding.RecoveryTrustBundleSHA256 = value.NewRecoveryTrustBundleSHA256
	binding.SignerStatus = "active"
	store.bindings[value.BindingID] = binding
	return nil
}

func (store *memoryStore) FenceSigner(
	_ context.Context,
	value ports.SignerFenceRecord,
) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	binding, found := store.bindings[value.BindingID]
	if !found {
		return ports.ErrNotFound
	}
	if binding.SignerGeneration != value.ExpectedSignerGeneration ||
		binding.SignerKID != value.ExpectedSignerKID {
		return ports.ErrStaleHead
	}
	if binding.SignerStatus == "rekey_required" {
		return nil
	}
	if binding.SignerStatus != "active" {
		return ports.ErrConflict
	}
	binding.SignerStatus = "rekey_required"
	store.bindings[value.BindingID] = binding
	return nil
}
