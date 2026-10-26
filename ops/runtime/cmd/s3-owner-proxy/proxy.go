package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ownerProxy struct {
	owner string
	proxy *httputil.ReverseProxy
}

func newOwnerProxy(upstream *url.URL, owner string) (*ownerProxy, error) {
	if upstream == nil || upstream.Scheme != "http" || upstream.Host == "" ||
		!validOwner(owner) {
		return nil, fmt.Errorf("invalid S3 owner proxy configuration")
	}
	result := &ownerProxy{owner: owner}
	result.proxy = httputil.NewSingleHostReverseProxy(upstream)
	result.proxy.ErrorHandler = func(writer http.ResponseWriter, _ *http.Request, err error) {
		http.Error(writer, fmt.Sprintf("S3 upstream unavailable: %v", err),
			http.StatusBadGateway)
	}
	return result, nil
}

func (proxy *ownerProxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet && request.URL.Path == "/livez" {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("ok"))
		return
	}
	provided := request.Header.Get("x-amz-expected-bucket-owner")
	if provided != "" && provided != proxy.owner {
		writeOwnerMismatch(writer)
		return
	}
	if request.Method == http.MethodGet && request.URL.Query().Has("acl") {
		writePrivateACL(writer, proxy.owner)
		return
	}
	proxy.proxy.ServeHTTP(writer, request)
}

func writePrivateACL(writer http.ResponseWriter, owner string) {
	document := privateACL(owner)
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(document)))
	writer.Header().Set("Content-Type", "application/xml")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(document)
}

func validOwner(value string) bool {
	if len(value) != 12 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func writeOwnerMismatch(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/xml")
	writer.WriteHeader(http.StatusForbidden)
	_, _ = writer.Write([]byte("<Error><Code>AccessDenied</Code>" +
		"<Message>Expected bucket owner mismatch</Message></Error>"))
}
