package actorbridge

import (
	"net/http"
	"strings"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

var strippedActorHeaders = []string{
	"Authorization", "Cookie", "Proxy-Authorization", "Proxy-Authenticate",
	actorassertionv1.HeaderName, "Forwarded", "Via", "X-Forwarded-For",
	"X-Forwarded-Host", "X-Forwarded-Proto", "X-Forwarded-Port", "X-Real-IP",
	"Connection", "Proxy-Connection", "Keep-Alive", "TE", "Trailer",
	"Transfer-Encoding", "Upgrade",
}

func stripActorBridgeHeaders(header http.Header) {
	connectionValues := append([]string(nil), header.Values("Connection")...)
	for _, name := range strippedActorHeaders {
		deleteHeaderFold(header, name)
	}
	for _, value := range connectionValues {
		for _, name := range strings.Split(value, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				deleteHeaderFold(header, name)
			}
		}
	}
	for name := range header {
		if strings.HasPrefix(strings.ToLower(name), "x-forwarded-") {
			delete(header, name)
		}
	}
}

func deleteHeaderFold(header http.Header, target string) {
	for name := range header {
		if strings.EqualFold(name, target) {
			delete(header, name)
		}
	}
}

func hasHeaderFold(header http.Header, target string) bool {
	for name := range header {
		if strings.EqualFold(name, target) {
			return true
		}
	}
	return false
}
