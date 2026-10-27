package application

import (
	"context"
	"encoding/base64"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *Coordinator) replaySignerRotation(
	ctx context.Context,
	request api.SignerRotationRequest,
	state artifact.CanonicalState,
	existing ports.OperationRecord,
) (api.PreparedAuthority, error) {
	rotation := existing.Proposal.Value.SignerRotation
	if existing.Purpose != "proposal_signer_rotation" || rotation == nil ||
		rotation.NewSignerPublicKeySHA256 != request.NewSignerPublicKeySHA256 ||
		!rotationBundleMatchesRequest(rotation, request) {
		return api.PreparedAuthority{}, api.ErrConflict
	}
	command := operationCommand{operationID: request.OperationID,
		purpose: "proposal_signer_rotation", state: state,
		generation:   existing.Proposal.Value.Generation,
		previousHead: existing.Proposal.Value.PreviousHeadSHA256,
		rotation:     rotation, issuedAt: request.IssuedAt,
		policy: artifactExpiry(request.Policy)}
	operation, err := service.prepareOperation(ctx, command)
	if err != nil {
		return api.PreparedAuthority{}, err
	}
	return preparedAuthority(operation), nil
}

func rotationBundleMatchesRequest(
	rotation *artifact.SignerRotationEvidence,
	request api.SignerRotationRequest,
) bool {
	if rotation.RecoveryTrustBundleBytes == nil {
		return len(request.RecoveryTrustBundle) == 0
	}
	return *rotation.RecoveryTrustBundleBytes ==
		base64.RawURLEncoding.EncodeToString(request.RecoveryTrustBundle)
}
