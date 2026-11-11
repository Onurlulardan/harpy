package apidocs

import (
	"fmt"
	"net/http"
	"strings"

	scalargo "github.com/bdpiprava/scalar-go"
)

type referenceHandler struct{}

func (referenceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin, err := referenceOrigin(r)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	sources, err := referenceSources(origin)
	if err != nil {
		http.Error(w, "API reference is unavailable", http.StatusInternalServerError)
		return
	}
	html, err := scalargo.NewV2(
		scalargo.WithSpecURL(sources[0].URL),
		scalargo.WithMultipleSources(sources...),
		scalargo.WithMetaDataOpts(scalargo.WithTitle("Harpy API")),
		scalargo.WithPersistAuth(true),
		scalargo.WithHideModels(),
		scalargo.WithHideAllClients(),
		scalargo.WithAuthenticationOpts(
			scalargo.WithPreferredSecurityScheme("bearerAuth"),
			scalargo.WithSecurityScheme("bearerAuth", scalargo.BearerScheme("")),
		),
		scalargo.WithOverrideCSS(referenceSignInCSS),
		scalargo.WithCustomHeadJS(referenceSignInHeadJS),
		scalargo.WithCustomBodyJS(referenceSignInBodyJS),
	)
	if err != nil {
		http.Error(w, "API reference is unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

func referenceOrigin(r *http.Request) (string, error) {
	if r == nil || r.Host == "" || strings.ContainsAny(r.Host, " /\t\r\n") {
		return "", fmt.Errorf("request host is invalid")
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	switch r.Header.Get("X-Forwarded-Proto") {
	case "https", "http":
		scheme = r.Header.Get("X-Forwarded-Proto")
	}
	return scheme + "://" + r.Host, nil
}
