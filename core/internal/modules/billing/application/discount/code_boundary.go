package discount

import discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"

func ParseCode(value string) (discountdomain.Code, error) {
	return discountdomain.NormalizeCode(value)
}

func CanonicalCode(value string) (string, error) {
	code, err := ParseCode(value)
	if err != nil {
		return "", err
	}
	return string(code), nil
}
