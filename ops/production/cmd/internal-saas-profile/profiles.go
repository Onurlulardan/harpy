package main

import (
	"fmt"
	"sort"
	"strings"
)

const (
	modeKey     = "DEPLOYMENT_MODE"
	registryKey = "HARPY_SAAS_AUTHORITY_REGISTRY_URL"
	proposerKey = "HARPY_SAAS_AUTHORITY_PROPOSER_KMS_KEY_ID"
	ackTrustKey = "HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH"
	wifKey      = "HARPY_SAAS_AUTHORITY_WORKLOAD_IDENTITY_CONFIG_PATH"
)

func renderProfiles(options profileOptions) (map[string][]byte, error) {
	common := map[string]string{modeKey: "saas", ackTrustKey: options.ackTrustBundle}
	api := cloneEnvironment(common)
	api[proposerKey] = options.proposerKMSKeyID
	worker := cloneEnvironment(common)
	worker[registryKey] = options.registryURL
	operator := cloneEnvironment(common)
	operator[registryKey] = options.registryURL
	operator[proposerKey] = options.proposerKMSKeyID
	if options.workloadIdentity != "" {
		api[wifKey] = options.workloadIdentity
		worker[wifKey] = options.workloadIdentity
		operator[wifKey] = options.workloadIdentity
	}
	environments := map[string]map[string]string{
		"api.saas.env":                 api,
		"worker.saas.env":              worker,
		"operator-initialize.saas.env": operator,
	}
	files := make(map[string][]byte, len(environments))
	for name, environment := range environments {
		content, err := renderEnvironment(environment)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", name, err)
		}
		files[name] = content
	}
	return files, nil
}

func cloneEnvironment(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func renderEnvironment(environment map[string]string) ([]byte, error) {
	keys := make([]string, 0, len(environment))
	for key, value := range environment {
		if !validEnvironmentKey(key) || !validScalar(value, 4096) {
			return nil, fmt.Errorf("invalid environment entry %q", key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var output strings.Builder
	for _, key := range keys {
		output.WriteString(key)
		output.WriteByte('=')
		output.WriteString(environment[key])
		output.WriteByte('\n')
	}
	return []byte(output.String()), nil
}

func validEnvironmentKey(value string) bool {
	if value == "" || value[0] < 'A' || value[0] > 'Z' {
		return false
	}
	for _, character := range value[1:] {
		if (character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}
