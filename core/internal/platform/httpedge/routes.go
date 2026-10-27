package httpedge

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type Routes struct {
	routes []compiledRoute
	report ErrorSink
}

func NewRoutes(report ErrorSink, entries ...Route) (*Routes, error) {
	if report == nil {
		return nil, fmt.Errorf("HTTP route error sink is required")
	}
	compiled := make([]compiledRoute, 0, len(entries))
	for _, entry := range entries {
		route, err := compileRoute(entry)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, route)
	}
	if err := validateRouteSet(compiled); err != nil {
		return nil, err
	}
	sort.Slice(compiled, func(left, right int) bool {
		if compiled[left].precedence != compiled[right].precedence {
			return compiled[left].precedence > compiled[right].precedence
		}
		if compiled[left].specificity != compiled[right].specificity {
			return compiled[left].specificity > compiled[right].specificity
		}
		if compiled[left].pattern != compiled[right].pattern {
			return compiled[left].pattern < compiled[right].pattern
		}
		return compiled[left].method < compiled[right].method
	})
	return &Routes{routes: compiled, report: report}, nil
}

func validateRouteSet(routes []compiledRoute) error {
	for left := range routes {
		for right := left + 1; right < len(routes); right++ {
			if routes[left].pattern == routes[right].pattern &&
				routes[left].method == routes[right].method {
				return fmt.Errorf("HTTP route is duplicated")
			}
			if routes[left].precedence == routes[right].precedence &&
				routes[left].specificity == routes[right].specificity &&
				routesOverlap(routes[left], routes[right]) &&
				routes[left].pattern != routes[right].pattern {
				return fmt.Errorf(
					"HTTP route match is ambiguous: %s %s overlaps %s %s",
					routes[left].method, routes[left].pattern,
					routes[right].method, routes[right].pattern,
				)
			}
		}
	}
	return nil
}

func (routes *Routes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	best := routes.bestMatches(r.URL.Path)
	for _, match := range best {
		if match.route.method != r.Method {
			continue
		}
		for index, segment := range match.route.segments {
			if segment.parameter != "" {
				r.SetPathValue(segment.parameter, match.values[index])
			}
		}
		*r = *r.WithContext(requestmeta.WithRoutePattern(r.Context(), match.route.pattern))
		match.route.handler.ServeHTTP(w, r)
		return
	}
	if len(best) == 0 {
		routes.writeFailure(w, r, httpwire.FailureRouteNotFound)
		return
	}
	*r = *r.WithContext(requestmeta.WithRoutePattern(r.Context(), best[0].route.pattern))
	allowed := make([]string, 0, len(best))
	for _, match := range best {
		allowed = append(allowed, match.route.method)
	}
	sort.Strings(allowed)
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	routes.writeFailure(w, r, httpwire.FailureMethodNotAllowed)
}

func (routes *Routes) bestMatches(requestPath string) []routeMatch {
	bestPrecedence := -1
	bestSpecificity := -1
	var matches []routeMatch
	for _, route := range routes.routes {
		match, found := matchRoute(route, requestPath)
		if !found {
			continue
		}
		if bestSpecificity < 0 {
			bestPrecedence = route.precedence
			bestSpecificity = route.specificity
		}
		if route.precedence < bestPrecedence ||
			(route.precedence == bestPrecedence && route.specificity < bestSpecificity) {
			break
		}
		matches = append(matches, match)
	}
	return matches
}

func (routes *Routes) writeFailure(
	w http.ResponseWriter,
	r *http.Request,
	failure httpwire.Failure,
) {
	correlationID, found := requestmeta.CorrelationIDFromContext(r.Context())
	if !found {
		correlationID = requestmeta.EntropyFailureCorrelationID
	}
	if err := httpwire.WriteFailure(
		w, correlationID, failure, r.Method == http.MethodHead,
	); err != nil {
		routes.report(err)
	}
}
