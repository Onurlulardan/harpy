package postgres

import (
	"context"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	"github.com/google/uuid"
)

func (database *ActorBridgeDatabase) ResolveIssueAuthority(
	ctx context.Context,
	tenantID uuid.UUID,
	domain actorassertionv1.Domain,
) (actorbridge.Authority, error) {
	if database == nil || actorBridgeQueryerIsNil(database.queryer) || ctx == nil ||
		ctx.Err() != nil || tenantID == uuid.Nil || domain.Validate() != nil {
		return actorbridge.Authority{}, actorbridge.NewFault(actorbridge.FaultInvalid)
	}
	var groupGeneration, iamGeneration, targetGeneration int64
	err := database.queryer.QueryRow(ctx, issueIAMV1ActorAssertionSQL,
		tenantID, string(domain)).Scan(&groupGeneration, &iamGeneration, &targetGeneration)
	if err != nil {
		return actorbridge.Authority{}, classifyActorBridgeError(err)
	}
	authority, err := actorbridge.NewAuthority(groupGeneration, iamGeneration, targetGeneration)
	if err != nil {
		return actorbridge.Authority{}, actorbridge.NewFault(actorbridge.FaultUnavailable)
	}
	return authority, nil
}
