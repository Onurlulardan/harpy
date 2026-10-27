package primarytarget

import tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"

type Queries struct {
	reader Reader
}

func New(reader Reader) (*Queries, error) {
	if reader == nil {
		return nil, tenantapi.ErrPrimaryTenantTargetUnavailable
	}
	return &Queries{reader: reader}, nil
}
