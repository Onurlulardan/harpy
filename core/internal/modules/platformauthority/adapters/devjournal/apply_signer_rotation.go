package devjournal

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

func (registry *Registry) applySignerRotation(
	ctx context.Context,
	state *journalState,
	proposal artifact.SignedProposal,
) error {
	evidence := proposal.Value.SignerRotation
	binding := state.bindings[proposal.Value.RegistryBindingID]
	if evidence == nil || evidence.RecoveryTrustBundleBytes == nil ||
		evidence.PreviousSignerGeneration != binding.signerGeneration ||
		proposal.Value.SignerKID != binding.signerKID {
		return fmt.Errorf("development authority signer rotation is stale")
	}
	document, err := base64.RawURLEncoding.DecodeString(*evidence.RecoveryTrustBundleBytes)
	if err != nil {
		return fmt.Errorf("development authority signer bundle is invalid")
	}
	bundle, err := recoveryv1.Parse(document)
	if err != nil || bundle.EnvironmentID() != binding.proof.Value.EnvironmentID ||
		bundle.Generation() != binding.signerGeneration+1 ||
		bundle.SHA256() != *proposal.Value.RecoveryTrustBundleSHA256 {
		return fmt.Errorf("development authority signer bundle is invalid")
	}
	newKID, kidFound := bundle.KeyID(recoveryv1.PurposeRecoveryControl)
	newPublic, publicFound := bundle.PublicKey(recoveryv1.PurposeRecoveryControl)
	newDigest := sha256.Sum256(newPublic[:])
	if !kidFound || !publicFound || newKID != evidence.NewSignerKID ||
		hex.EncodeToString(newDigest[:]) != evidence.NewSignerPublicKeySHA256 ||
		!bundle.Retains(recoveryv1.PurposeRecoveryControl, binding.signerKID,
			binding.signerPublicKeySHA256) {
		return fmt.Errorf("development authority signer rotation conflicts")
	}
	resolved, err := registry.resolver.ResolveProposalKey(ctx, evidence.NewSignerKID)
	resolvedDigest := sha256.Sum256(resolved)
	if err != nil || string(resolved) != string(newPublic[:]) ||
		hex.EncodeToString(resolvedDigest[:]) != evidence.NewSignerPublicKeySHA256 {
		return fmt.Errorf("development authority new signer key is unavailable")
	}
	binding.signerGeneration++
	binding.signerKID = evidence.NewSignerKID
	binding.signerPublicKeySHA256 = evidence.NewSignerPublicKeySHA256
	state.bindings[proposal.Value.RegistryBindingID] = binding
	return nil
}
