package billing

import (
	"fmt"
	"strings"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func parseProviderJobReference(
	value string,
	prefix string,
) (string, uuid.UUID, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 3 || parts[0] != prefix {
		return "", uuid.Nil, fmt.Errorf("billing provider job reference is invalid")
	}
	provider, err := providerapp.CanonicalProviderName(parts[1])
	if err != nil || provider != parts[1] {
		return "", uuid.Nil, fmt.Errorf("billing provider job reference is invalid")
	}
	identifier, err := uuid.Parse(parts[2])
	if err != nil || identifier == uuid.Nil {
		return "", uuid.Nil, fmt.Errorf("billing provider job reference is invalid")
	}
	return provider, identifier, nil
}
