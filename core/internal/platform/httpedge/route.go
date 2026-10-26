package httpedge

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

const (
	maximumRoutePatternBytes = 256
	maximumRouteSegments     = 32
	maximumRoutePrecedence   = 100
)

type Route struct {
	Method     string
	Pattern    string
	Handler    http.Handler
	Precedence int
}

type routeSegment struct {
	literal   string
	parameter string
}

type compiledRoute struct {
	method      string
	pattern     string
	handler     http.Handler
	segments    []routeSegment
	specificity int
	precedence  int
}

func compileRoute(entry Route) (compiledRoute, error) {
	if !validRouteMethod(entry.Method) || entry.Handler == nil ||
		!validRoutePattern(entry.Pattern) || entry.Precedence < 0 ||
		entry.Precedence > maximumRoutePrecedence {
		return compiledRoute{}, fmt.Errorf("HTTP route is invalid")
	}
	parts := splitRoutePath(entry.Pattern)
	segments := make([]routeSegment, 0, len(parts))
	parameters := make(map[string]struct{})
	specificity := 0
	for _, part := range parts {
		segment, err := compileRouteSegment(part)
		if err != nil {
			return compiledRoute{}, err
		}
		if segment.parameter != "" {
			if _, duplicate := parameters[segment.parameter]; duplicate {
				return compiledRoute{}, fmt.Errorf("HTTP route parameter is duplicated")
			}
			parameters[segment.parameter] = struct{}{}
		} else {
			specificity++
		}
		segments = append(segments, segment)
	}
	return compiledRoute{
		method: entry.Method, pattern: entry.Pattern, handler: entry.Handler,
		segments: segments, specificity: specificity, precedence: entry.Precedence,
	}, nil
}

func validRouteMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}

func validRoutePattern(pattern string) bool {
	return pattern != "" && len(pattern) <= maximumRoutePatternBytes &&
		strings.HasPrefix(pattern, "/") && !strings.Contains(pattern, "\\") &&
		!strings.ContainsAny(pattern, "?#") && path.Clean(pattern) == pattern &&
		len(splitRoutePath(pattern)) <= maximumRouteSegments
}

func splitRoutePath(value string) []string {
	if value == "/" {
		return nil
	}
	return strings.Split(strings.TrimPrefix(value, "/"), "/")
}

func compileRouteSegment(value string) (routeSegment, error) {
	if len(value) >= 3 && value[0] == '{' && value[len(value)-1] == '}' {
		parameter := value[1 : len(value)-1]
		if validRouteParameter(parameter) {
			return routeSegment{parameter: parameter}, nil
		}
		return routeSegment{}, fmt.Errorf("HTTP route parameter is invalid")
	}
	if !validRouteLiteral(value) {
		return routeSegment{}, fmt.Errorf("HTTP route literal is invalid")
	}
	return routeSegment{literal: value}, nil
}

func validRouteParameter(value string) bool {
	if value == "" || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value[1:] {
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func validRouteLiteral(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') &&
			!strings.ContainsRune("-._~", character) {
			return false
		}
	}
	return true
}
