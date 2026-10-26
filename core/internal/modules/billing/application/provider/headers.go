package providerapp

import (
	"net/textproto"
	"regexp"
	"strings"
)

const (
	maximumHeaderFields = 64
	maximumHeaderBytes  = 32 << 10
)

var headerNamePattern = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$")

type Headers struct {
	values map[string][]string
}

func NewHeaders(source map[string][]string) (Headers, error) {
	if len(source) == 0 || len(source) > maximumHeaderFields {
		return Headers{}, ErrWebhookRequest
	}
	values := make(map[string][]string, len(source))
	total := 0
	for name, entries := range source {
		if !headerNamePattern.MatchString(name) || len(entries) == 0 {
			return Headers{}, ErrWebhookRequest
		}
		canonical := textproto.CanonicalMIMEHeaderKey(name)
		for _, entry := range entries {
			total += len(canonical) + len(entry)
			if entry == "" || strings.ContainsAny(entry, "\r\n") || total > maximumHeaderBytes {
				return Headers{}, ErrWebhookRequest
			}
			values[canonical] = append(values[canonical], entry)
		}
	}
	return Headers{values: values}, nil
}

func (headers Headers) Values(name string) []string {
	if !headerNamePattern.MatchString(name) {
		return nil
	}
	return append([]string(nil), headers.values[textproto.CanonicalMIMEHeaderKey(name)]...)
}

func cloneHeaders(headers Headers) Headers {
	copy := make(map[string][]string, len(headers.values))
	for name, values := range headers.values {
		copy[name] = append([]string(nil), values...)
	}
	return Headers{values: copy}
}
