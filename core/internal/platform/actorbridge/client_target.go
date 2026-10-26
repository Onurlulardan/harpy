package actorbridge

import (
	"net"
	"net/url"
	"slices"
	"strings"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

var exactTargetDomains = []actorassertionv1.Domain{
	actorassertionv1.DomainAnalytics, actorassertionv1.DomainBilling,
	actorassertionv1.DomainCatalog, actorassertionv1.DomainDocuments,
	actorassertionv1.DomainForms, actorassertionv1.DomainNotifications,
	actorassertionv1.DomainWork, actorassertionv1.DomainWorkflow,
}

type Target struct {
	domain     actorassertionv1.Domain
	baseURL    *url.URL
	serverName string
}

func NewTarget(
	domain actorassertionv1.Domain,
	baseURL string,
	serverName string,
) (Target, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || domain.Validate() != nil || parsed.Scheme != "https" ||
		parsed.User != nil || parsed.Host == "" || parsed.Path != "" ||
		parsed.RawPath != "" || parsed.RawQuery != "" || parsed.Fragment != "" ||
		parsed.String() != baseURL || serverName != parsed.Hostname() ||
		!validServerName(serverName) {
		return Target{}, NewFault(FaultInvalid)
	}
	return Target{domain: domain, baseURL: parsed, serverName: serverName}, nil
}

func (target Target) Domain() actorassertionv1.Domain { return target.domain }

func validServerName(value string) bool {
	if len(value) < 1 || len(value) > 253 || net.ParseIP(value) != nil ||
		strings.HasPrefix(value, ".") || strings.HasSuffix(value, ".") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if (character < 'a' || character > 'z') &&
				(character < '0' || character > '9') && character != '-' {
				return false
			}
		}
	}
	return true
}

func exactTargets(targets map[actorassertionv1.Domain]Target) bool {
	actual := make([]string, 0, len(targets))
	for domain, target := range targets {
		if target.domain != domain || target.baseURL == nil || target.serverName == "" {
			return false
		}
		actual = append(actual, string(domain))
	}
	expected := make([]string, len(exactTargetDomains))
	for index, domain := range exactTargetDomains {
		expected[index] = string(domain)
	}
	slices.Sort(actual)
	slices.Sort(expected)
	return slices.Equal(actual, expected)
}
