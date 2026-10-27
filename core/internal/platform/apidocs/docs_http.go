package apidocs

import (
	_ "embed"
	"fmt"
	"net/http"
)

//go:embed index.html
var indexHTML []byte

//go:embed app.js
var applicationJS []byte

//go:embed style.css
var stylesheet []byte

type Handler struct {
	enabled  bool
	document []byte
	etag     string
}

func New(enabled bool, document []byte) (*Handler, error) {
	if !enabled {
		return &Handler{}, nil
	}
	validated, etag, err := validateDocument(document)
	if err != nil {
		return nil, fmt.Errorf("build API documentation handler: %w", err)
	}
	return &Handler{enabled: true, document: validated, etag: etag}, nil
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	setSecurityHeaders(w.Header())
	if handler == nil || !handler.enabled || request == nil || request.URL == nil {
		http.NotFound(w, request)
		return
	}
	if !knownDocumentationPath(request.URL.Path) {
		http.NotFound(w, request)
		return
	}
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	switch request.URL.Path {
	case "/swagger":
		http.Redirect(w, request, "/swagger/", http.StatusPermanentRedirect)
	case "/swagger/", "/swagger/index.html":
		writeAsset(w, request, "text/html; charset=utf-8", indexHTML)
	case "/swagger/app.js":
		writeAsset(w, request, "text/javascript; charset=utf-8", applicationJS)
	case "/swagger/style.css":
		writeAsset(w, request, "text/css; charset=utf-8", stylesheet)
	case "/swagger/openapi.json":
		handler.writeDocument(w, request)
	}
}

func knownDocumentationPath(path string) bool {
	switch path {
	case "/swagger", "/swagger/", "/swagger/index.html", "/swagger/app.js",
		"/swagger/style.css", "/swagger/openapi.json":
		return true
	default:
		return false
	}
}

func (handler *Handler) writeDocument(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("ETag", handler.etag)
	if request.Header.Get("If-None-Match") == handler.etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	writeAsset(w, request, "application/vnd.oai.openapi+json", handler.document)
}

func writeAsset(w http.ResponseWriter, request *http.Request, contentType string, content []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Length", fmt.Sprint(len(content)))
	w.WriteHeader(http.StatusOK)
	if request.Method != http.MethodHead {
		_, _ = w.Write(content)
	}
}

func setSecurityHeaders(header http.Header) {
	header.Set("Cache-Control", "no-store")
	header.Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Content-Type-Options", "nosniff")
}
