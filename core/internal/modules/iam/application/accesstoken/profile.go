package accesstoken

import (
	"net"
	"net/url"
	"strings"
)

type Profile struct {
	issuer   string
	audience string
}

func NewProfile(issuer, audience string, allowLocalHTTP bool) (Profile, error) {
	if !validEndpoint(issuer, allowLocalHTTP) ||
		!validEndpoint(audience, allowLocalHTTP) {
		return Profile{}, ErrInvalidProfile
	}
	return Profile{issuer: issuer, audience: audience}, nil
}

func (profile Profile) Issuer() string   { return profile.issuer }
func (profile Profile) Audience() string { return profile.audience }

func validEndpoint(value string, allowLocalHTTP bool) bool {
	if value == "" || strings.TrimSpace(value) != value {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" ||
		parsed.User != nil || parsed.Fragment != "" || parsed.RawQuery != "" ||
		parsed.String() != value {
		return false
	}
	if parsed.Scheme == "https" {
		return true
	}
	return allowLocalHTTP && parsed.Scheme == "http" && localHost(parsed.Hostname())
}

func localHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
