package onboarding

import (
	"context"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (storage *postgresStore) ReceiveAuthorityAck(
	ctx context.Context,
	ack tenantapi.AuthorityAcknowledgement,
	expectedRevision int64,
) (tenantapi.OnboardingTransition, error) {
	if !storage.usable(ctx) || !ack.IsValid() || expectedRevision < 1 {
		return tenantapi.OnboardingTransition{}, errUnavailable
	}
	transition, err := scanOnboardingTransition(storage.queryRow(ctx,
		recordAuthorityAckSQL, uuid.UUID(ack.OperationID), uuid.UUID(ack.RegistryBindingID),
		string(ack.ScopeKind), uuid.UUID(ack.ScopeID), ack.Generation,
		optionalDigest(ack.PreviousHeadSHA256), ack.StateSHA256, ack.ProposalSHA256,
		ack.AckSHA256, ack.RegistrySequence, ack.VerifiedAt, expectedRevision))
	if err != nil {
		return tenantapi.OnboardingTransition{}, fmt.Errorf("receive authority ack row: %w", err)
	}
	return transition, nil
}
