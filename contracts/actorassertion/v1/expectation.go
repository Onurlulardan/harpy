package actorassertionv1

import "github.com/google/uuid"

type Expectation struct {
	TenantID                                            uuid.UUID
	TargetDomain                                        Domain
	ResourceKey, ActionKey                              string
	TargetBindingSHA256                                 [32]byte
	IAMGroupGeneration, IAMGeneration, TargetGeneration int64
}

func (expectation Expectation) Validate() error {
	if expectation.TenantID == uuid.Nil || expectation.TargetDomain.Validate() != nil ||
		!validCatalogKey(expectation.ResourceKey) || !validCatalogKey(expectation.ActionKey) ||
		expectation.TargetBindingSHA256 == [32]byte{} ||
		expectation.IAMGroupGeneration <= 0 || expectation.IAMGeneration <= 0 ||
		expectation.TargetGeneration <= 0 {
		return ErrContract
	}
	return nil
}

func (expectation Expectation) matches(claims Claims) bool {
	decision := claims.Decision
	return expectation.Validate() == nil && decision.TenantID == expectation.TenantID &&
		decision.TargetDomain == expectation.TargetDomain &&
		decision.ResourceKey == expectation.ResourceKey && decision.ActionKey == expectation.ActionKey &&
		decision.TargetBindingSHA256 == expectation.TargetBindingSHA256 &&
		decision.IAMGroupGeneration == expectation.IAMGroupGeneration &&
		decision.IAMGeneration == expectation.IAMGeneration &&
		decision.TargetGeneration == expectation.TargetGeneration
}
