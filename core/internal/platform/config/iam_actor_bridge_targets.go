package config

import (
	"encoding/json"
	"io"
	"net"
	"net/url"
	"slices"
	"strings"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

var iamActorBridgeDomains = []actorassertionv1.Domain{
	actorassertionv1.DomainAnalytics,
	actorassertionv1.DomainBilling,
	actorassertionv1.DomainCatalog,
	actorassertionv1.DomainDocuments,
	actorassertionv1.DomainForms,
	actorassertionv1.DomainNotifications,
	actorassertionv1.DomainWork,
	actorassertionv1.DomainWorkflow,
}

type IAMActorBridgeTarget struct {
	domain     actorassertionv1.Domain
	baseURL    string
	serverName string
}

type actorBridgeTargetsDocument struct {
	Version int                         `json:"version"`
	Targets []actorBridgeTargetDocument `json:"targets"`
}

type actorBridgeTargetDocument struct {
	Domain     string `json:"domain"`
	BaseURL    string `json:"base_url"`
	ServerName string `json:"server_name"`
}

func loadActorBridgeTargets(
	values environmentValues,
) (map[string]IAMActorBridgeTarget, error) {
	path, found := values.lookup(iamActorBridgeTargetsFileKey)
	if !found {
		return nil, configError(iamActorBridgeTargetsFileKey, "is required")
	}
	document, err := readSecretFile(iamActorBridgeTargetsFileKey, path, secretFileLimit)
	if err != nil {
		return nil, err
	}
	if rejectDuplicateJSONMembers([]byte(document)) != nil {
		return nil, configError(iamActorBridgeTargetsFileKey, "contains ambiguous JSON")
	}
	decoder := json.NewDecoder(newStringReader(document))
	decoder.DisallowUnknownFields()
	var raw actorBridgeTargetsDocument
	if decoder.Decode(&raw) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		raw.Version != 1 || len(raw.Targets) != len(iamActorBridgeDomains) {
		return nil, configError(iamActorBridgeTargetsFileKey, "must contain the exact V1 domain set")
	}
	targets := make(map[string]IAMActorBridgeTarget, len(raw.Targets))
	for _, entry := range raw.Targets {
		target, valid := newIAMActorBridgeTarget(entry)
		if !valid {
			return nil, configError(iamActorBridgeTargetsFileKey, "contains an invalid target")
		}
		key := string(target.domain)
		if _, duplicate := targets[key]; duplicate {
			return nil, configError(iamActorBridgeTargetsFileKey, "contains a duplicate domain")
		}
		targets[key] = target
	}
	if !exactActorBridgeDomains(targets) {
		return nil, configError(iamActorBridgeTargetsFileKey, "must contain the exact V1 domain set")
	}
	return targets, nil
}

func newIAMActorBridgeTarget(
	document actorBridgeTargetDocument,
) (IAMActorBridgeTarget, bool) {
	domain := actorassertionv1.Domain(document.Domain)
	parsed, err := url.Parse(document.BaseURL)
	if err != nil || domain.Validate() != nil || parsed.Scheme != "https" ||
		parsed.User != nil || parsed.Host == "" || parsed.Path != "" ||
		parsed.RawPath != "" || parsed.RawQuery != "" || parsed.Fragment != "" ||
		parsed.String() != document.BaseURL || document.ServerName != parsed.Hostname() ||
		!validActorBridgeServerName(document.ServerName) {
		return IAMActorBridgeTarget{}, false
	}
	return IAMActorBridgeTarget{domain: domain, baseURL: document.BaseURL,
		serverName: document.ServerName}, true
}

func validActorBridgeServerName(value string) bool {
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

func exactActorBridgeDomains(targets map[string]IAMActorBridgeTarget) bool {
	actual := make([]string, 0, len(targets))
	for domain := range targets {
		actual = append(actual, domain)
	}
	slices.Sort(actual)
	expected := make([]string, len(iamActorBridgeDomains))
	for index, domain := range iamActorBridgeDomains {
		expected[index] = string(domain)
	}
	slices.Sort(expected)
	return slices.Equal(actual, expected)
}
