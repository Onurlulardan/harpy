package application

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func buildRotationEvidence(
	request api.SignerRotationRequest,
	binding ports.BindingRecord,
) (*artifact.SignerRotationEvidence, bool) {
	if !validDigest(request.NewSignerPublicKeySHA256) ||
		request.NewSignerPublicKeySHA256 == binding.SignerPublicKeySHA256 ||
		request.State.ProposalSignerGeneration != binding.SignerGeneration+1 ||
		request.State.ProposalSignerKID == binding.SignerKID {
		return nil, false
	}
	evidence := &artifact.SignerRotationEvidence{
		PreviousSignerGeneration: binding.SignerGeneration,
		NewSignerKID:             request.State.ProposalSignerKID,
		NewSignerPublicKeySHA256: request.NewSignerPublicKeySHA256,
	}
	if request.State.DeploymentMode == api.ModeSaaS {
		return evidence, len(request.RecoveryTrustBundle) == 0 &&
			request.State.RecoveryTrustBundleSHA256 == nil
	}
	bundle, err := recoveryv1.Parse(request.RecoveryTrustBundle)
	if err != nil || request.State.RecoveryTrustBundleSHA256 == nil ||
		bundle.EnvironmentID() != request.State.EnvironmentID ||
		bundle.Generation() != request.State.ProposalSignerGeneration ||
		bundle.SHA256() != *request.State.RecoveryTrustBundleSHA256 {
		return nil, false
	}
	controlKID, found := bundle.KeyID(recoveryv1.PurposeRecoveryControl)
	controlPublic, publicFound := bundle.PublicKey(recoveryv1.PurposeRecoveryControl)
	publicDigest := sha256.Sum256(controlPublic[:])
	if !found || !publicFound || controlKID != request.State.ProposalSignerKID ||
		hex.EncodeToString(publicDigest[:]) != request.NewSignerPublicKeySHA256 ||
		!bundle.Retains(recoveryv1.PurposeRecoveryControl, binding.SignerKID,
			binding.SignerPublicKeySHA256) {
		return nil, false
	}
	encoded := base64.RawURLEncoding.EncodeToString(bundle.CanonicalBytes())
	evidence.RecoveryTrustBundleBytes = &encoded
	return evidence, true
}

func validDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && hex.EncodeToString(decoded) == value
}
