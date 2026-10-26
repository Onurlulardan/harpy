package config

import (
	"encoding/json"
	"net"
	"net/url"
	"strings"
)

const (
	iamAccessIssuerEnvironmentKey   = "HARPY_IAM_ACCESS_TOKEN_ISSUER"
	iamAccessAudienceEnvironmentKey = "HARPY_IAM_ACCESS_TOKEN_AUDIENCE"
	iamAllowedOriginsEnvironmentKey = "HARPY_IAM_ALLOWED_ORIGINS"
	maximumIAMOrigins               = 32
	maximumIAMEndpointBytes         = 512
)

type IAMHTTPProfile struct {
	issuer         string
	audience       string
	origins        []string
	allowLocalHTTP bool
}

func loadIAMHTTPProfile(
	values environmentValues,
	environment Environment,
) (IAMHTTPProfile, error) {
	issuer, found := values.lookup(iamAccessIssuerEnvironmentKey)
	if !found {
		return IAMHTTPProfile{}, configError(iamAccessIssuerEnvironmentKey, "is required")
	}
	audience, found := values.lookup(iamAccessAudienceEnvironmentKey)
	if !found {
		return IAMHTTPProfile{}, configError(iamAccessAudienceEnvironmentKey, "is required")
	}
	rawOrigins, found := values.lookup(iamAllowedOriginsEnvironmentKey)
	if !found {
		return IAMHTTPProfile{}, configError(iamAllowedOriginsEnvironmentKey, "is required")
	}
	allowLocalHTTP := environment == EnvironmentDevelopment || environment == EnvironmentTest
	if !validIAMEndpoint(issuer, allowLocalHTTP) {
		return IAMHTTPProfile{}, configError(iamAccessIssuerEnvironmentKey, "must be a canonical HTTPS URI")
	}
	if !validIAMEndpoint(audience, allowLocalHTTP) {
		return IAMHTTPProfile{}, configError(iamAccessAudienceEnvironmentKey, "must be a canonical HTTPS URI")
	}
	origins, err := parseIAMOrigins(rawOrigins, allowLocalHTTP)
	if err != nil {
		return IAMHTTPProfile{}, err
	}
	return IAMHTTPProfile{
		issuer: issuer, audience: audience, origins: origins, allowLocalHTTP: allowLocalHTTP,
	}, nil
}

func (profile IAMHTTPProfile) Validate() error {
	if !validIAMEndpoint(profile.issuer, profile.allowLocalHTTP) ||
		!validIAMEndpoint(profile.audience, profile.allowLocalHTTP) ||
		len(profile.origins) == 0 || len(profile.origins) > maximumIAMOrigins {
		return errInvalidIAMAPIConfig
	}
	seen := make(map[string]struct{}, len(profile.origins))
	for _, origin := range profile.origins {
		if !validIAMOrigin(origin, profile.allowLocalHTTP) {
			return errInvalidIAMAPIConfig
		}
		if _, duplicate := seen[origin]; duplicate {
			return errInvalidIAMAPIConfig
		}
		seen[origin] = struct{}{}
	}
	return nil
}

func (profile IAMHTTPProfile) Issuer() string       { return profile.issuer }
func (profile IAMHTTPProfile) Audience() string     { return profile.audience }
func (profile IAMHTTPProfile) AllowLocalHTTP() bool { return profile.allowLocalHTTP }
func (profile IAMHTTPProfile) AllowedOrigins() []string {
	return append([]string(nil), profile.origins...)
}

func parseIAMOrigins(raw string, allowLocalHTTP bool) ([]string, error) {
	if raw == "" || len(raw) > maximumIAMOrigins*(maximumIAMEndpointBytes+4) {
		return nil, configError(iamAllowedOriginsEnvironmentKey, "must be a bounded JSON string array")
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	var origins []string
	if err := decoder.Decode(&origins); err != nil || origins == nil {
		return nil, configError(iamAllowedOriginsEnvironmentKey, "must be a JSON string array")
	}
	if decoder.Decode(&struct{}{}) == nil || len(origins) == 0 || len(origins) > maximumIAMOrigins {
		return nil, configError(iamAllowedOriginsEnvironmentKey, "must contain one bounded document")
	}
	canonical, err := json.Marshal(origins)
	if err != nil || string(canonical) != raw {
		return nil, configError(iamAllowedOriginsEnvironmentKey, "must use canonical JSON encoding")
	}
	seen := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		if !validIAMOrigin(origin, allowLocalHTTP) {
			return nil, configError(iamAllowedOriginsEnvironmentKey, "contains a noncanonical origin")
		}
		if _, duplicate := seen[origin]; duplicate {
			return nil, configError(iamAllowedOriginsEnvironmentKey, "contains a duplicate origin")
		}
		seen[origin] = struct{}{}
	}
	return append([]string(nil), origins...), nil
}

func validIAMEndpoint(value string, allowLocalHTTP bool) bool {
	parsed, ok := parseIAMURI(value)
	return ok && validIAMScheme(parsed, allowLocalHTTP)
}

func validIAMOrigin(value string, allowLocalHTTP bool) bool {
	parsed, ok := parseIAMURI(value)
	return ok && parsed.Path == "" && validIAMScheme(parsed, allowLocalHTTP)
}

func parseIAMURI(value string) (*url.URL, bool) {
	if value == "" || len(value) > maximumIAMEndpointBytes || strings.TrimSpace(value) != value {
		return nil, false
	}
	parsed, err := url.Parse(value)
	return parsed, err == nil && parsed.IsAbs() && parsed.Host != "" && parsed.User == nil &&
		!parsed.ForceQuery && parsed.Opaque == "" && parsed.RawQuery == "" &&
		parsed.Fragment == "" && parsed.String() == value
}

func validIAMScheme(parsed *url.URL, allowLocalHTTP bool) bool {
	if parsed.Scheme == "https" {
		return true
	}
	return allowLocalHTTP && parsed.Scheme == "http" && iamLoopback(parsed.Hostname())
}

func iamLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
