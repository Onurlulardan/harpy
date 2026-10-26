package iamactorv1

import (
	"context"
	"errors"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func (consumer *Consumer) Consume(
	ctx context.Context,
	transaction Transaction,
	wire string,
	expectation actorassertionv1.Expectation,
	effect Effect,
) (Receipt, error) {
	if !consumer.Ready() || ctx == nil || ctx.Err() != nil ||
		nilTransaction(transaction) || effect == nil {
		return Receipt{}, newError(CodeInvalid)
	}
	claims, err := consumer.verifier.Verify(wire, expectation, consumer.now())
	if err != nil {
		if errors.Is(err, actorassertionv1.ErrExpired) {
			return Receipt{}, newError(CodeExpired)
		}
		return Receipt{}, newError(CodeUnauthorized)
	}
	receipt, err := consumeVerified(ctx, transaction, claims)
	if err != nil {
		return Receipt{}, err
	}
	if err := effect(ctx, transaction, claims); err != nil {
		return Receipt{}, err
	}
	if !receipt.Valid() {
		return Receipt{}, newError(CodeUnavailable)
	}
	return receipt, nil
}

func consumeVerified(
	ctx context.Context,
	transaction Transaction,
	claims actorassertionv1.Claims,
) (Receipt, error) {
	if ctx == nil || ctx.Err() != nil || nilTransaction(transaction) ||
		claims.Validate() != nil {
		return Receipt{}, newError(CodeInvalid)
	}
	decision := claims.Decision
	var consumedAt time.Time
	err := transaction.QueryRow(ctx, consumeStatement,
		decision.TenantID, claims.JTI, decision.SubjectID, decision.PermissionID,
		string(decision.TargetDomain), actorassertionv1.Audience(decision.TargetDomain),
		claims.KeyID, claims.DecisionSHA256[:], decision.TargetBindingSHA256[:],
		decision.IAMGroupGeneration, decision.IAMGeneration, decision.TargetGeneration,
		decision.TokenVersion, decision.AuthorizationVersion, decision.AuthorizationClock,
		claims.IssuedAt, claims.ExpiresAt,
	).Scan(&consumedAt)
	if err != nil {
		return Receipt{}, classifyDatabaseError(err)
	}
	return newReceipt(consumedAt)
}
