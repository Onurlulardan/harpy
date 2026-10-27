package transaction

type ProviderScopeRequest struct {
	Lookup   ProviderLookup
	Resolver ProviderResolverBinder
}

func (request ProviderScopeRequest) valid() bool {
	return request.Lookup.valid() && request.Resolver.valid
}
