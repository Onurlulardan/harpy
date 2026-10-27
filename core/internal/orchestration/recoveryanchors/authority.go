package recoveryanchors

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	platformauthorityapi "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/google/uuid"
)

type Authority struct {
	planner platformauthorityapi.ProtectedAuthorityPlanner
}

func NewAuthority(
	planner platformauthorityapi.ProtectedAuthorityPlanner,
) (*Authority, error) {
	if planner == nil {
		return nil, fmt.Errorf("recovery anchor authority planner is required")
	}
	return &Authority{planner: planner}, nil
}

func (authority *Authority) Prepare(
	ctx context.Context,
	operationID string,
	state anchorapi.AuthorityState,
	now time.Time,
) (anchorapi.PreparedAuthority, error) {
	if authority == nil || authority.planner == nil ||
		state.DeploymentMode != "onpremise" || !validAuthorityUUID(operationID) ||
		!validAuthorityState(state) {
		return anchorapi.PreparedAuthority{}, anchorapi.ErrInvalid
	}
	ownerID, ownerGeneration := state.OwnerBindingID, state.OwnerBindingGeneration
	primary := state.PrimaryAdminBindingSetSHA256
	anchors := state.RecoveryAnchorSetSHA256
	trust := state.RecoveryTrustBundleSHA256
	prepared, err := authority.planner.PrepareChange(ctx,
		platformauthorityapi.ProtectedChangeRequest{
			OperationID: operationID, Purpose: platformauthorityapi.ChangeRecoveryAnchor,
			State: platformauthorityapi.State{
				Schema:                   "protected-authority-state.v1",
				EnvironmentID:            state.EnvironmentID,
				DeploymentMode:           platformauthorityapi.ModeOnPremise,
				RegistryBindingID:        state.RegistryBindingID,
				ScopeKind:                platformauthorityapi.ScopeEnvironment,
				ScopeID:                  state.EnvironmentID,
				ProposalSignerGeneration: state.ProposalSignerGeneration,
				ProposalSignerKID:        state.ProposalSignerKID,
				OwnerBindingID:           &ownerID, OwnerBindingGeneration: &ownerGeneration,
				PrimaryAdminBindingSetSHA256: &primary,
				RecoveryAnchorSetSHA256:      &anchors,
				RecoveryTrustBundleSHA256:    &trust,
			}, IssuedAt: now, Policy: platformauthorityapi.ExpiryOnline,
		})
	if err != nil {
		if errors.Is(err, platformauthorityapi.ErrConflict) ||
			errors.Is(err, platformauthorityapi.ErrStaleHead) ||
			errors.Is(err, platformauthorityapi.ErrBindingConflict) {
			return anchorapi.PreparedAuthority{}, anchorapi.ErrConflict
		}
		return anchorapi.PreparedAuthority{}, anchorapi.ErrUnavailable
	}
	return anchorapi.PreparedAuthority{OperationID: prepared.OperationID,
		RegistryBindingID: prepared.RegistryBindingID, Generation: prepared.Generation,
		StateSHA256: prepared.StateSHA256, ProposalSHA256: prepared.ProposalSHA256}, nil
}

func validAuthorityState(state anchorapi.AuthorityState) bool {
	return validAuthorityUUID(state.EnvironmentID) &&
		validAuthorityUUID(state.OwnerBindingID) &&
		validAuthorityUUID(state.RegistryBindingID) &&
		state.OwnerBindingGeneration > 0 && state.ProposalSignerGeneration > 0 &&
		state.ProposalSignerKID != "" &&
		validAuthorityDigest(state.PrimaryAdminBindingSetSHA256) &&
		validAuthorityDigest(state.RecoveryAnchorSetSHA256) &&
		validAuthorityDigest(state.RecoveryTrustBundleSHA256)
}

func validAuthorityUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func validAuthorityDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	valid := err == nil && len(decoded) == 32 && hex.EncodeToString(decoded) == value
	clear(decoded)
	return valid
}

var _ anchorapi.AuthorityPlanner = (*Authority)(nil)
