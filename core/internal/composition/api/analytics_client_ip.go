package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func newAnalyticsClientIPResolver(
	proxies config.AnalyticsTrustedProxies,
) analyticsmodule.HTTPTrustedClientIPResolver {
	return func(request *http.Request) (string, error) {
		if request == nil || request.Header == nil || forwardedHeaderConflict(request.Header) {
			return "", fmt.Errorf("Analytics client IP evidence is invalid")
		}
		peer, err := remoteIP(request.RemoteAddr)
		if err != nil {
			return "", err
		}
		forwarded := request.Header.Values("X-Forwarded-For")
		if !proxies.Contains(peer) {
			if len(forwarded) != 0 {
				return "", fmt.Errorf("Analytics forwarding peer is not trusted")
			}
			return peer.String(), nil
		}
		if len(forwarded) == 0 {
			return peer.String(), nil
		}
		if len(forwarded) != 1 {
			return "", fmt.Errorf("Analytics forwarding chain is ambiguous")
		}
		chain, err := parseForwardedChain(forwarded[0])
		if err != nil {
			return "", err
		}
		for index := len(chain) - 1; index >= 0; index-- {
			if !proxies.Contains(chain[index]) {
				return chain[index].String(), nil
			}
		}
		return chain[0].String(), nil
	}
}

func forwardedHeaderConflict(header http.Header) bool {
	for _, name := range []string{"Forwarded", "Via", "X-Real-IP",
		"X-Forwarded-Host", "X-Forwarded-Proto", "X-Forwarded-Port"} {
		if len(header.Values(name)) != 0 {
			return true
		}
	}
	return false
}

func remoteIP(address string) (net.IP, error) {
	host, _, err := net.SplitHostPort(address)
	ip := net.ParseIP(host)
	if err != nil || ip == nil || ip.IsUnspecified() {
		return nil, fmt.Errorf("Analytics peer address is invalid")
	}
	return ip, nil
}

func parseForwardedChain(value string) ([]net.IP, error) {
	parts := strings.Split(value, ",")
	if len(parts) == 0 || len(parts) > 16 {
		return nil, fmt.Errorf("Analytics forwarding chain is invalid")
	}
	result := make([]net.IP, len(parts))
	for index, part := range parts {
		wire := strings.TrimSpace(part)
		ip := net.ParseIP(wire)
		if wire == "" || ip == nil || ip.IsUnspecified() || ip.String() != wire {
			return nil, fmt.Errorf("Analytics forwarding address is invalid")
		}
		result[index] = ip
	}
	return result, nil
}
