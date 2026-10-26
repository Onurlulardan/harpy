package work

import (
	"strings"

	workhttp "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/http"
)

const compatibilityContractPrefix = "work-v1-http.v1:"

func validHTTPAccess(access HTTPAccess, operation, resource string,
	compatibility bool,
) bool {
	contract, found := workhttp.ExpectedAccess(operation)
	return found && contract.Resource == resource &&
		contract.Compatibility == compatibility &&
		access.ValidFor(contract.Authority, operation, compatibility)
}

func compatibilityOperation(hit HTTPCompatibilityHit) (string, bool) {
	if !strings.HasPrefix(hit.RouteContract, compatibilityContractPrefix) {
		return "", false
	}
	operation := strings.TrimPrefix(hit.RouteContract, compatibilityContractPrefix)
	return operation, operation != "" && hit.ConsumerID == hit.Access.ConsumerID() &&
		hit.BuildID == hit.Access.BuildID()
}
