package api

import (
	"fmt"
	"net/http"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func buildPublicRoutes(
	security *iamHTTPSecurity,
	report httpedge.ErrorSink,
	groups ...[]httpedge.Route,
) (*httpedge.Routes, error) {
	if security == nil || !security.Valid() {
		return nil, fmt.Errorf("IAM HTTP security is required")
	}
	jwks, err := security.NewJWKSHandler(iammodule.HTTPErrorSink(report))
	if err != nil {
		return nil, fmt.Errorf("build JWKS handler: %w", err)
	}
	entries := []httpedge.Route{
		httpedge.Route{
			Method: http.MethodGet, Pattern: "/.well-known/jwks.json", Handler: jwks,
		},
		httpedge.Route{
			Method: http.MethodHead, Pattern: "/.well-known/jwks.json", Handler: jwks,
		},
	}
	for _, group := range groups {
		entries = append(entries, group...)
	}
	routes, err := httpedge.NewRoutes(report, entries...)
	if err != nil {
		return nil, fmt.Errorf("build public routes: %w", err)
	}
	return routes, nil
}
