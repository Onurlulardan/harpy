package identifier

import (
	"strings"
	"unicode/utf8"
)

const (
	maximumEmailLocalBytes = 64
	maximumEmailBytes      = 254
)

func NormalizeEmailV1(raw string) (Value, error) {
	if !utf8.ValidString(raw) {
		return Value{}, ErrInvalid
	}
	display := strings.Trim(raw, " \t\r\n")
	if strings.Count(display, "@") != 1 {
		return Value{}, ErrInvalid
	}
	local, domain, _ := strings.Cut(display, "@")
	if invalidEmailLocal(local) || invalidEmailDomainInput(domain) {
		return Value{}, ErrInvalid
	}
	local = unicodeDefaultCasefoldV1(local)
	if invalidEmailLocal(local) || len([]byte(local)) > maximumEmailLocalBytes {
		return Value{}, ErrInvalid
	}
	domain, err := normalizeEmailDomainV1(domain)
	if err != nil {
		return Value{}, ErrInvalid
	}
	normalized := local + "@" + domain
	if len([]byte(normalized)) > maximumEmailBytes {
		return Value{}, ErrInvalid
	}
	return newValue(EmailV1, display, normalized), nil
}

func invalidEmailLocal(local string) bool {
	return local == "" || strings.HasPrefix(local, ".") ||
		strings.HasSuffix(local, ".") || strings.Contains(local, "..") ||
		strings.ContainsAny(local, `"()[]\`) || containsASCIISpaceOrControl(local)
}

func containsASCIISpaceOrControl(value string) bool {
	for _, character := range value {
		if character <= 0x20 || character == 0x7f {
			return true
		}
	}
	return false
}
