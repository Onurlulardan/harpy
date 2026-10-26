package workhttp

type AccessContract struct {
	Authority, Resource string
	Compatibility       bool
}

func ExpectedAccess(operation string) (AccessContract, bool) {
	var result AccessContract
	found := false
	for _, route := range append(canonicalRoutes(), compatibilityRoutes()...) {
		if route.operation != operation {
			continue
		}
		candidate := AccessContract{Authority: route.authority,
			Resource: route.resource, Compatibility: route.compatibility}
		if found && candidate != result {
			return AccessContract{}, false
		}
		result, found = candidate, true
	}
	return result, found
}
