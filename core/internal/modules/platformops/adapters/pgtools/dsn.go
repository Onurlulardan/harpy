package pgtools

import (
	"errors"
	"net/url"
	"sort"
	"strings"
)

var allowedQuery = map[string]bool{
	"connect_timeout":      true,
	"sslmode":              true,
	"sslrootcert":          true,
	"target_session_attrs": true,
}

func sanitizedDSN(raw string) (string, string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed == nil || parsed.Scheme != "postgresql" ||
		parsed.Host == "" || parsed.Path == "" || parsed.Fragment != "" || parsed.User == nil {
		return "", "", errors.New("PostgreSQL DSN is invalid")
	}
	password, found := parsed.User.Password()
	if !found || password == "" || strings.ContainsAny(password, "\x00\r\n") {
		return "", "", errors.New("PostgreSQL DSN password is invalid")
	}
	username := parsed.User.Username()
	if username == "" {
		return "", "", errors.New("PostgreSQL DSN principal is invalid")
	}
	query := parsed.Query()
	keys := make([]string, 0, len(query))
	for key, values := range query {
		if !allowedQuery[key] || len(values) != 1 || values[0] == "" {
			return "", "", errors.New("PostgreSQL DSN option is invalid")
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	cleanQuery := make(url.Values, len(keys))
	for _, key := range keys {
		cleanQuery.Set(key, query.Get(key))
	}
	parsed.User = url.User(username)
	parsed.RawQuery = cleanQuery.Encode()
	return parsed.String(), password, nil
}
