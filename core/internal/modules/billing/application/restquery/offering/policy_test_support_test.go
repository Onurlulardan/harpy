package offering

import restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"

func zeroGlobalRequest() restauthorization.GlobalPolicyRequest {
	return restauthorization.GlobalPolicyRequest{}
}
