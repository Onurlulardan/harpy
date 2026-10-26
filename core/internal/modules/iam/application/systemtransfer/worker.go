package systemtransfer

type Worker struct {
	scopes      FreshBatchScopes
	authorizer  BatchAuthorizer
	identities  IdentityLookup
	resources   ResourceMutator
	invitations InvitationCommand
	receipts    ReceiptStore
}

func NewWorker(
	scopes FreshBatchScopes,
	authorizer BatchAuthorizer,
	identities IdentityLookup,
	resources ResourceMutator,
	invitations InvitationCommand,
	receipts ReceiptStore,
) (*Worker, error) {
	if nilDependency(scopes) || nilDependency(authorizer) ||
		nilDependency(identities) || nilDependency(resources) ||
		nilDependency(invitations) || nilDependency(receipts) {
		return nil, ErrDependencies
	}
	return &Worker{scopes: scopes, authorizer: authorizer, identities: identities,
		resources: resources, invitations: invitations, receipts: receipts}, nil
}
