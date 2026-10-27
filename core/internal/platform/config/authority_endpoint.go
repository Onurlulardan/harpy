package config

import (
	"net/url"
	"regexp"
	"strings"
)

type ServiceURL struct{ value string }

func (value ServiceURL) Value() string { return value.value }

func loadServiceURL(
	values environmentValues,
	key string,
	environment Environment,
) (ServiceURL, error) {
	raw, found := values.lookup(key)
	if !found {
		return ServiceURL{}, configError(key, "is required")
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" ||
		parsed.Fragment != "" || strings.TrimSpace(raw) != raw {
		return ServiceURL{}, configError(key, "must be an absolute canonical HTTP URL")
	}
	if parsed.Scheme != "https" &&
		(parsed.Scheme != "http" || environment.productionLike()) {
		return ServiceURL{}, configError(key, "must use an allowed HTTP scheme")
	}
	return ServiceURL{value: raw}, nil
}

type AuthorityKMSKeyID struct{ value string }

func (key AuthorityKMSKeyID) Value() string { return key.value }

var authorityKMSKeyIDPattern = regexp.MustCompile(
	`^gcp-kms://projects/[a-z][a-z0-9-]{4,62}/locations/[a-z0-9-]+/` +
		`keyRings/[A-Za-z0-9_-]+/cryptoKeys/[A-Za-z0-9_-]+/cryptoKeyVersions/[1-9][0-9]*$`,
)

func loadAuthorityKMSKeyID(
	values environmentValues,
	environment Environment,
) (AuthorityKMSKeyID, error) {
	raw, found := values.lookup(saasKMSKeyIDKey)
	if !found {
		if environment.productionLike() {
			return AuthorityKMSKeyID{}, configError(
				saasKMSKeyIDKey, "must identify one versioned GCP KMS asymmetric signing key",
			)
		}
		return AuthorityKMSKeyID{}, nil
	}
	if !authorityKMSKeyIDPattern.MatchString(raw) {
		return AuthorityKMSKeyID{}, configError(
			saasKMSKeyIDKey, "must identify one versioned GCP KMS asymmetric signing key",
		)
	}
	return AuthorityKMSKeyID{value: raw}, nil
}

func rejectConfigured(values environmentValues, keys ...string) error {
	for _, key := range keys {
		if _, found := values.lookup(key); found {
			return configError(key, "is forbidden for the selected profile mode")
		}
	}
	return nil
}
