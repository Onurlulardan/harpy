package config

import (
	"net"
	"strings"
)

const analyticsTrustedProxyKey = "HARPY_ANALYTICS_TRUSTED_PROXY_CIDRS"

type AnalyticsTrustedProxies struct{ networks []*net.IPNet }

func loadAnalyticsTrustedProxies(
	values environmentValues,
) (AnalyticsTrustedProxies, error) {
	raw := values.valueOr(analyticsTrustedProxyKey, "")
	if raw == "" {
		return AnalyticsTrustedProxies{}, nil
	}
	if strings.TrimSpace(raw) != raw {
		return AnalyticsTrustedProxies{}, configError(
			analyticsTrustedProxyKey, "must not have surrounding whitespace")
	}
	parts := strings.Split(raw, ",")
	if len(parts) > 16 {
		return AnalyticsTrustedProxies{}, configError(
			analyticsTrustedProxyKey, "must contain at most 16 networks")
	}
	result := AnalyticsTrustedProxies{networks: make([]*net.IPNet, 0, len(parts))}
	for _, part := range parts {
		ip, network, err := net.ParseCIDR(part)
		if err != nil || ip.String() == "<nil>" || network.String() != part ||
			part == "0.0.0.0/0" || part == "::/0" {
			return AnalyticsTrustedProxies{}, configError(
				analyticsTrustedProxyKey, "must contain canonical bounded CIDRs")
		}
		for _, existing := range result.networks {
			if existing.Contains(ip) || network.Contains(existing.IP) {
				return AnalyticsTrustedProxies{}, configError(
					analyticsTrustedProxyKey, "must not contain overlapping networks")
			}
		}
		result.networks = append(result.networks, network)
	}
	return result, nil
}

func (proxies AnalyticsTrustedProxies) Contains(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, network := range proxies.networks {
		if network != nil && network.Contains(ip) {
			return true
		}
	}
	return false
}

func (proxies AnalyticsTrustedProxies) Count() int { return len(proxies.networks) }
