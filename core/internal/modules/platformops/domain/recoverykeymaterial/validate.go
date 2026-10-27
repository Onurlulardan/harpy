package recoverykeymaterial

import (
	"regexp"
	"strings"
)

var keyIDPattern = regexp.MustCompile(`^[a-z][a-z0-9._:-]{0,127}$`)

func validKeyID(value string) bool {
	return keyIDPattern.MatchString(value)
}

func validProvider(kind, handle string) bool {
	return (kind == ProviderLocalKeyring || kind == ProviderForgeGuard) &&
		len(handle) >= 1 && len(handle) <= 512 &&
		!strings.ContainsAny(handle, "\x00\r\n")
}
