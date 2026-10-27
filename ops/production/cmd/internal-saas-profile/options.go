package main

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var proposerKMSKeyIDPattern = regexp.MustCompile(
	`^gcp-kms://projects/[a-z][a-z0-9-]{4,62}/locations/[a-z0-9-]+/` +
		`keyRings/[A-Za-z0-9_-]+/cryptoKeys/[A-Za-z0-9_-]+/` +
		`cryptoKeyVersions/[1-9][0-9]*$`,
)

type profileOptions struct {
	output           string
	registryURL      string
	proposerKMSKeyID string
	ackTrustBundle   string
	workloadIdentity string
}

func (options profileOptions) validate() (profileOptions, error) {
	resolved, err := filepath.Abs(filepath.Clean(options.output))
	if err != nil || filepath.Dir(resolved) == resolved {
		return profileOptions{}, fmt.Errorf("output must be a dedicated directory")
	}
	options.output = resolved
	if !canonicalHTTPSURL(options.registryURL) {
		return profileOptions{}, fmt.Errorf("authority registry URL is invalid")
	}
	if !proposerKMSKeyIDPattern.MatchString(options.proposerKMSKeyID) {
		return profileOptions{}, fmt.Errorf("proposer KMS key ID is invalid")
	}
	if !canonicalContainerPath(options.ackTrustBundle) {
		return profileOptions{}, fmt.Errorf("ack trust bundle path is invalid")
	}
	if options.workloadIdentity != "" &&
		!canonicalContainerPath(options.workloadIdentity) {
		return profileOptions{}, fmt.Errorf("workload identity path is invalid")
	}
	return options, nil
}

func canonicalHTTPSURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" &&
		parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" &&
		parsed.String() == value
}

func validScalar(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && strings.TrimSpace(value) == value &&
		!strings.ContainsAny(value, "\x00\r\n")
}

func canonicalContainerPath(value string) bool {
	return validScalar(value, 4096) && path.IsAbs(value) &&
		path.Clean(value) == value && value != "/"
}
