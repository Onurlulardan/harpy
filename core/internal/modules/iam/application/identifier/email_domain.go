package identifier

import (
	"strings"

	"golang.org/x/net/idna"
)

const maximumEmailDomainBytes = 253

var emailDomainProfileV1 = idna.New(
	idna.MapForLookup(),
	idna.Transitional(false),
	idna.StrictDomainName(true),
	idna.ValidateLabels(true),
	idna.VerifyDNSLength(true),
	idna.BidiRule(),
	idna.CheckHyphens(true),
	idna.CheckJoiners(true),
)

func invalidEmailDomainInput(domain string) bool {
	return domain == "" || strings.HasPrefix(domain, ".") ||
		strings.HasSuffix(domain, ".") || strings.Contains(domain, "..") ||
		strings.ContainsAny(domain, `"()[]\`) || containsASCIISpaceOrControl(domain)
}

func normalizeEmailDomainV1(domain string) (string, error) {
	ascii, err := emailDomainProfileV1.ToASCII(domain)
	if err != nil {
		return "", ErrInvalid
	}
	ascii = strings.ToLower(ascii)
	if len(ascii) == 0 || len(ascii) > maximumEmailDomainBytes {
		return "", ErrInvalid
	}
	for _, label := range strings.Split(ascii, ".") {
		if len(label) < 1 || len(label) > 63 {
			return "", ErrInvalid
		}
	}
	return ascii, nil
}
