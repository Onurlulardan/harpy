package transaction

type CredentialScopeRequest struct {
	Lookup   CredentialLookup
	Resolver CredentialResolverBinder
}

func (request CredentialScopeRequest) valid() bool {
	return request.Lookup.valid() && request.Resolver.valid
}
