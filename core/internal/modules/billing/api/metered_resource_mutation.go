package billingapi

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type MeteredResourceMutationRequest struct {
	Resource            MeteredResourceKind
	Mutation            MeteredResourceMutation
	TenantID            [16]byte
	ActorID             [16]byte
	ProposedOperationID [16]byte
	ProposedOwnerFactID [16]byte
	IdempotencyKey      string
	RequestSHA256       [32]byte
}

func (request MeteredResourceMutationRequest) Valid() bool {
	return request.Resource.Valid() && request.Mutation.Valid() &&
		request.TenantID != ([16]byte{}) && request.ActorID != ([16]byte{}) &&
		validUUIDv7(request.ProposedOperationID) &&
		validUUIDv7(request.ProposedOwnerFactID) &&
		validIdempotencyKey(request.IdempotencyKey) &&
		request.RequestSHA256 != ([32]byte{})
}

type MeteredResourceMutationDecision struct {
	Resource       MeteredResourceKind
	Mutation       MeteredResourceMutation
	TenantID       [16]byte
	ActorID        [16]byte
	OperationID    [16]byte
	OwnerFactID    [16]byte
	IdempotencyKey string
	RequestSHA256  [32]byte
	FeatureEnabled bool
	Allowed        bool
	Replay         bool
	GaugeDelta     int64
	Current        int64
	Limit          int64
	Unlimited      bool
}

func (decision MeteredResourceMutationDecision) ValidFor(
	request MeteredResourceMutationRequest,
) bool {
	if !request.Valid() || decision.Resource != request.Resource ||
		decision.Mutation != request.Mutation || decision.TenantID != request.TenantID ||
		decision.ActorID != request.ActorID ||
		decision.IdempotencyKey != request.IdempotencyKey ||
		decision.RequestSHA256 != request.RequestSHA256 ||
		!validUUIDv7(decision.OperationID) || !validUUIDv7(decision.OwnerFactID) ||
		decision.Current < 0 || decision.Limit < 0 ||
		(decision.Unlimited && decision.Limit != 0) {
		return false
	}
	if !decision.Replay && (decision.OperationID != request.ProposedOperationID ||
		decision.OwnerFactID != request.ProposedOwnerFactID) {
		return false
	}
	if decision.Replay && request.Mutation != MeteredResourceCreate &&
		decision.OwnerFactID != request.ProposedOwnerFactID {
		return false
	}
	return decision.validOutcome()
}

func (decision MeteredResourceMutationDecision) validOutcome() bool {
	switch decision.Mutation {
	case MeteredResourceUpdate, MeteredResourcePublish, MeteredResourceRevalidate:
		return decision.Allowed == decision.FeatureEnabled &&
			decision.GaugeDelta == 0 && decision.Current == 0 &&
			decision.Limit == 0 && !decision.Unlimited
	case MeteredResourceCreate, MeteredResourceRestore:
		capacity := decision.Unlimited || decision.Current < decision.Limit
		wantDelta := int64(0)
		if decision.Allowed {
			wantDelta = 1
		}
		return decision.Allowed == (decision.FeatureEnabled && capacity) &&
			decision.GaugeDelta == wantDelta
	case MeteredResourceDelete:
		return decision.Allowed && !decision.FeatureEnabled &&
			(decision.GaugeDelta == -1 || decision.GaugeDelta == 0)
	default:
		return false
	}
}

type MeteredResourceMutationDecider interface {
	DecideMeteredResourceMutation(
		context.Context,
		platformtx.TenantScope,
		MeteredResourceMutationRequest,
	) (MeteredResourceMutationDecision, error)
}

func validUUIDv7(value [16]byte) bool {
	return value != ([16]byte{}) && value[6]>>4 == 7 && value[8]&0xc0 == 0x80
}

func validIdempotencyKey(value string) bool {
	if len(value) < 1 || len(value) > 128 {
		return false
	}
	for index := range value {
		character := value[index]
		if character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' ||
			index > 0 && (character == '.' || character == '_' ||
				character == ':' || character == '-') {
			continue
		}
		return false
	}
	return true
}
