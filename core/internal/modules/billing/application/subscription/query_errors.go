package subscriptionapp

import "errors"

var (
	ErrQueryConfiguration           = errors.New("billing subscription query is misconfigured")
	ErrQueryScope                   = errors.New("billing subscription query scope is invalid")
	ErrCurrentSubscriptionNotFound  = errors.New("current billing subscription does not exist")
	ErrCurrentSubscriptionInvariant = errors.New("current billing subscription invariant failed")
	ErrSubscriptionQueryDependency  = errors.New("billing subscription query dependency is unavailable")
	ErrSubscriptionQueryStorage     = errors.New("billing subscription query storage failed")
)
