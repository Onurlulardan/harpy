package foundation

import (
	"regexp"
	"sort"
	"strings"
)

var userinfoPattern = regexp.MustCompile(`([a-z][a-z0-9+.-]*://)[^/\s]+@`)

func Redact(message string, secrets []string) string {
	ordered := append([]string(nil), secrets...)
	sort.Slice(ordered, func(i, j int) bool {
		return len(ordered[i]) > len(ordered[j])
	})
	for _, secret := range ordered {
		if len(secret) < 12 {
			continue
		}
		message = strings.ReplaceAll(message, secret, "[redacted]")
	}
	return userinfoPattern.ReplaceAllString(message, `${1}[redacted]@`)
}
