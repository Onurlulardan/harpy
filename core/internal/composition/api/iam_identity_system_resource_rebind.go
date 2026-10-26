package api

import (
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func rebindIdentitySystemResourceRoutes(
	direct []httpedge.Route,
	canonical []httpedge.Route,
) ([]httpedge.Route, error) {
	expected := identitySystemResourceCRUDRouteKeys()
	if len(canonical) != len(expected) {
		return nil, fmt.Errorf("canonical system resource route inventory is incomplete")
	}
	canonicalFound := make(map[string]struct{}, len(expected))
	for _, route := range canonical {
		key := route.Method + "\x00" + route.Pattern
		if _, found := expected[key]; !found || route.Handler == nil {
			return nil, fmt.Errorf("canonical system resource route is invalid")
		}
		if _, duplicate := canonicalFound[key]; duplicate {
			return nil, fmt.Errorf("canonical system resource route is duplicated")
		}
		canonicalFound[key] = struct{}{}
	}
	found := make(map[string]struct{}, len(expected))
	preserved := make([]httpedge.Route, 0, len(direct)-len(expected))
	for _, route := range direct {
		key := route.Method + "\x00" + route.Pattern
		if _, replaced := expected[key]; replaced {
			if _, duplicate := found[key]; duplicate {
				return nil, fmt.Errorf("direct system resource route is duplicated")
			}
			found[key] = struct{}{}
			continue
		}
		preserved = append(preserved, route)
	}
	if len(found) != len(expected) {
		return nil, fmt.Errorf("direct system resource route inventory drifted")
	}
	return preserved, nil
}

func identitySystemResourceCRUDRouteKeys() map[string]struct{} {
	resources := []string{"users", "roles", "organizations"}
	result := make(map[string]struct{}, len(resources)*5)
	for _, resource := range resources {
		collection := "/api/" + resource
		item := collection + "/{id}"
		for _, route := range [][2]string{
			{http.MethodGet, collection},
			{http.MethodPost, collection},
			{http.MethodGet, item},
			{http.MethodPut, item},
			{http.MethodDelete, item},
		} {
			result[route[0]+"\x00"+route[1]] = struct{}{}
		}
	}
	return result
}
