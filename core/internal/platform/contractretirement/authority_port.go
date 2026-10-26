package contractretirement

import "context"

type AuthorityPort interface {
	Provision(
		context.Context, Key, Identity, RemainingScopeCount,
	) (Revision, error)
	Reconcile(
		context.Context, Key, Identity, RemainingScopeCount, Revision,
	) (Revision, error)
	MarkEligible(context.Context, Key, Identity, Revision) (Revision, error)
	Finalize(context.Context, Key, Identity, Revision) (Revision, error)
}
