package csrf

import (
	"net"
	"net/url"
	"strings"
)

type OriginPolicy struct{ allowed map[string]struct{} }

func NewOriginPolicy(origins []string, allowLocalHTTP bool) (OriginPolicy, error) {
	if len(origins) == 0 {
		return OriginPolicy{}, ErrInvalidPolicy
	}
	policy := OriginPolicy{allowed: make(map[string]struct{}, len(origins))}
	for _, origin := range origins {
		if !validOrigin(origin, allowLocalHTTP) {
			return OriginPolicy{}, ErrInvalidPolicy
		}
		if _, duplicate := policy.allowed[origin]; duplicate {
			return OriginPolicy{}, ErrInvalidPolicy
		}
		policy.allowed[origin] = struct{}{}
	}
	return policy, nil
}

func (policy OriginPolicy) Allows(origin string) bool {
	if len(policy.allowed) == 0 {
		return false
	}
	_, allowed := policy.allowed[origin]
	return allowed
}

func validOrigin(value string, allowLocalHTTP bool) bool {
	if value == "" || len(value) > MaximumOriginSize || strings.TrimSpace(value) != value ||
		strings.EqualFold(value, "null") {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" || parsed.User != nil ||
		parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" ||
		parsed.String() != value {
		return false
	}
	if parsed.Scheme == "https" {
		return true
	}
	return allowLocalHTTP && parsed.Scheme == "http" && isLoopback(parsed.Hostname())
}

func isLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
