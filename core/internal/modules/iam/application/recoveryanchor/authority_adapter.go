package recoveryanchor

import (
	"context"
	"encoding/hex"
	"time"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	anchorports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/recoveryanchor"
	"github.com/google/uuid"
)

type authorityAdapter struct{ planner anchorapi.AuthorityPlanner }

func NewAuthorityPlanner(planner anchorapi.AuthorityPlanner) (anchorports.AuthorityPlanner, error) {
	if planner == nil {
		return nil, anchordomain.ErrDependency
	}
	return authorityAdapter{planner: planner}, nil
}

func (adapter authorityAdapter) Prepare(
	ctx context.Context,
	operationID anchordomain.ID,
	state anchordomain.AuthorityState,
	now time.Time,
) (anchordomain.PreparedAuthority, error) {
	prepared, err := adapter.planner.Prepare(ctx, uuid.UUID(operationID).String(),
		anchorapi.AuthorityState{
			EnvironmentID:            uuid.UUID(state.EnvironmentID).String(),
			OwnerBindingID:           uuid.UUID(state.OwnerBindingID).String(),
			RegistryBindingID:        uuid.UUID(state.RegistryBindingID).String(),
			DeploymentMode:           state.DeploymentMode,
			ProposalSignerKID:        state.ProposalSignerKID,
			OwnerBindingGeneration:   state.OwnerBindingGeneration,
			ProposalSignerGeneration: state.ProposalSignerGeneration,
			PrimaryAdminBindingSetSHA256: hex.EncodeToString(
				state.PrimaryAdminBindingSetSHA256[:]),
			RecoveryAnchorSetSHA256:   hex.EncodeToString(state.RecoveryAnchorSetSHA256[:]),
			RecoveryTrustBundleSHA256: state.RecoveryTrustBundleSHA256,
		}, now)
	if err != nil {
		return anchordomain.PreparedAuthority{}, anchordomain.ErrDependency
	}
	return preparedAuthorityDomain(prepared)
}

func preparedAuthorityDomain(
	value anchorapi.PreparedAuthority,
) (anchordomain.PreparedAuthority, error) {
	operationID, operationErr := uuid.Parse(value.OperationID)
	registryID, registryErr := uuid.Parse(value.RegistryBindingID)
	stateSHA, stateErr := authorityDigest(value.StateSHA256)
	proposalSHA, proposalErr := authorityDigest(value.ProposalSHA256)
	if operationErr != nil || registryErr != nil || stateErr != nil ||
		proposalErr != nil || value.Generation < 2 {
		return anchordomain.PreparedAuthority{}, anchordomain.ErrDependency
	}
	return anchordomain.PreparedAuthority{OperationID: operationID,
		RegistryBindingID: registryID, Generation: value.Generation,
		StateSHA256: stateSHA, ProposalSHA256: proposalSHA}, nil
}

func authorityDigest(value string) ([32]byte, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 || hex.EncodeToString(decoded) != value {
		clear(decoded)
		return [32]byte{}, anchordomain.ErrDependency
	}
	var digest [32]byte
	copy(digest[:], decoded)
	clear(decoded)
	return digest, nil
}

var _ anchorports.AuthorityPlanner = authorityAdapter{}
