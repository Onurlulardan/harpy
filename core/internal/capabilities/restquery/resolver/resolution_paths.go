package resolver

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func indexedPath(base string, index int) string {
	return fmt.Sprintf("%s/%d", base, index)
}

func childFilterPath(base string, kind api.RequestFilterKind, index int) string {
	switch kind {
	case api.RequestFilterAnd:
		return indexedPath(base+"/and", index)
	case api.RequestFilterOr:
		return indexedPath(base+"/or", index)
	case api.RequestFilterNot:
		return base + "/not"
	default:
		return base
	}
}

func childHavingPath(base string, kind api.RequestHavingKind, index int) string {
	switch kind {
	case api.RequestHavingKindAnd:
		return indexedPath(base+"/and", index)
	case api.RequestHavingKindOr:
		return indexedPath(base+"/or", index)
	case api.RequestHavingKindNot:
		return base + "/not"
	default:
		return base
	}
}

func literalMemberPath(base string, operator api.Operator) string {
	arity, valid := operator.Arity()
	if valid && arity == api.PredicateArityOne {
		return base + "/value"
	}
	return base + "/values"
}

func mustRequestPath(value string) resolved.RequestPath {
	path, err := resolved.NewRequestPath(value)
	if err != nil {
		// All paths are assembled from bounded structural constants and indices.
		// Preserve fail-closed behavior if that internal invariant ever changes.
		return resolved.RequestPath{}
	}
	return path
}
