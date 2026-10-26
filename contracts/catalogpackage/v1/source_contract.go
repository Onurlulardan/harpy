package catalogpackagev1

import (
	"fmt"
	"regexp"
	"strings"
)

const FirstPartySourceContractPrefix = "harpy-v1.enterprise-seed."

var firstPartyModuleSlugPattern = regexp.MustCompile(
	`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`,
)

func FirstPartySourceContract(moduleSlug string) (string, error) {
	if len(moduleSlug) == 0 || len(moduleSlug) > 63 ||
		!firstPartyModuleSlugPattern.MatchString(moduleSlug) {
		return "", fmt.Errorf("first-party module slug is invalid")
	}
	return FirstPartySourceContractPrefix + moduleSlug, nil
}

func ParseFirstPartySourceContract(value string) (string, error) {
	moduleSlug, found := strings.CutPrefix(value, FirstPartySourceContractPrefix)
	if !found {
		return "", fmt.Errorf("first-party source contract prefix is invalid")
	}
	expected, err := FirstPartySourceContract(moduleSlug)
	if err != nil || expected != value {
		return "", fmt.Errorf("first-party source contract is invalid")
	}
	return moduleSlug, nil
}

func ValidFirstPartySourceContract(value string) bool {
	_, err := ParseFirstPartySourceContract(value)
	return err == nil
}
