//go:build integration

package postgres

import (
	"context"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
)

const actorBridgeTenantConsumeSQL = `
SELECT iam_v1_actor_assertion_consume_v1(
    current_schema()::name,
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17
)`

type actorBridgeTestConsumer struct{ queryer ActorBridgeQueryer }

func newActorBridgeTestConsumer(queryer ActorBridgeQueryer) *actorBridgeTestConsumer {
	return &actorBridgeTestConsumer{queryer: queryer}
}

func (consumer *actorBridgeTestConsumer) Consume(
	ctx context.Context, claims actorassertionv1.Claims,
) (actorbridge.Receipt, error) {
	decision := claims.Decision
	var consumedAt time.Time
	err := consumer.queryer.QueryRow(ctx, actorBridgeTenantConsumeSQL,
		decision.TenantID, claims.JTI, decision.SubjectID, decision.PermissionID,
		string(decision.TargetDomain), actorassertionv1.Audience(decision.TargetDomain),
		claims.KeyID, claims.DecisionSHA256[:], decision.TargetBindingSHA256[:],
		decision.IAMGroupGeneration, decision.IAMGeneration, decision.TargetGeneration,
		decision.TokenVersion, decision.AuthorizationVersion, decision.AuthorizationClock,
		claims.IssuedAt, claims.ExpiresAt,
	).Scan(&consumedAt)
	if err != nil {
		return actorbridge.Receipt{}, classifyActorBridgeError(err)
	}
	return actorbridge.NewReceipt(consumedAt)
}
