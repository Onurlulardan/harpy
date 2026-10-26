package config

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"
)

type PostgresEndpointKind string

const PostgresEndpointDirect PostgresEndpointKind = "direct_postgres"

func loadPostgresEndpointKind(values environmentValues) (PostgresEndpointKind, error) {
	value, found := values.lookup("HARPY_POSTGRES_ENDPOINT_KIND")
	if !found || value != string(PostgresEndpointDirect) {
		return "", configError(
			"HARPY_POSTGRES_ENDPOINT_KIND",
			"must be direct_postgres",
		)
	}
	return PostgresEndpointDirect, nil
}

func validatePostgresDSN(key string, secret Secret, environment Environment) error {
	parsed, err := url.Parse(secret.Value())
	if err != nil || parsed == nil {
		return configError(key, "must be a valid PostgreSQL URL")
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return configError(key, "must be a PostgreSQL URL")
	}
	if parsed.User == nil || parsed.User.Username() == "" || parsed.Hostname() == "" ||
		strings.Contains(parsed.Host, ",") || containsASCIIControl(parsed.Host) {
		return configError(key, "must include user and host")
	}
	password, _ := parsed.User.Password()
	if containsASCIIControl(parsed.User.Username()) || containsASCIIControl(password) {
		return configError(key, "contains an invalid credential encoding")
	}
	if parsed.Fragment != "" || parsed.Path == "" || parsed.Path == "/" || strings.Count(parsed.Path, "/") != 1 {
		return configError(key, "must include one database and no fragment")
	}
	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return configError(key, "contains an invalid query")
	}
	for name, values := range query {
		if len(values) != 1 || !allowedPostgresDSNParameter(name) {
			return configError(key, "contains an unsupported connection parameter")
		}
	}
	tlsMode := query.Get("sslmode")
	if environment.productionLike() && tlsMode != "verify-full" {
		return configError(key, "uses an invalid TLS mode for the environment")
	}
	if !environment.productionLike() && tlsMode != "disable" && tlsMode != "verify-full" {
		return configError(key, "uses an invalid TLS mode for the environment")
	}
	if err := validatePostgresTLSFiles(query, tlsMode); err != nil {
		return configError(key, err.Error())
	}
	return nil
}

func postgresDSNPrincipalAndDatabase(secret Secret) (string, string, error) {
	parsed, err := url.Parse(secret.Value())
	if err != nil || parsed == nil || parsed.User == nil {
		return "", "", errors.New("invalid PostgreSQL URL")
	}
	database, err := url.PathUnescape(strings.TrimPrefix(parsed.EscapedPath(), "/"))
	if err != nil || database == "" || strings.Contains(database, "/") {
		return "", "", errors.New("invalid PostgreSQL database")
	}
	return parsed.User.Username(), database, nil
}

func allowedPostgresDSNParameter(name string) bool {
	switch name {
	case "sslmode", "sslrootcert", "sslcert", "sslkey", "sslpassword":
		return true
	default:
		return false
	}
}

func validatePostgresTLSFiles(query url.Values, mode string) error {
	if mode == "disable" {
		if len(query) != 1 {
			return errors.New("disabled TLS mode cannot carry TLS options")
		}
		return nil
	}
	for _, key := range []string{"sslrootcert", "sslcert", "sslkey"} {
		value := query.Get(key)
		if value != "" && (!filepath.IsAbs(value) || filepath.Clean(value) != value) {
			return errors.New("TLS certificate paths must be absolute and clean")
		}
	}
	if query.Get("sslpassword") != "" && query.Get("sslkey") == "" {
		return errors.New("TLS sslpassword requires sslkey")
	}
	return nil
}

func containsASCIIControl(value string) bool {
	for _, character := range []byte(value) {
		if character < 0x20 || character == 0x7f {
			return true
		}
	}
	return false
}
