package deploymentpolicy

import tenancyapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"

type Queries struct {
	reader PolicyReader
}

func New(reader PolicyReader) (*Queries, error) {
	if reader == nil {
		return nil, tenancyapi.ErrDeploymentPolicyUnavailable
	}
	return &Queries{reader: reader}, nil
}
