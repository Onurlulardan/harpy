package systemresources

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func layoutResource(resource publicresource.Resource) (layoutapi.ResourceKind, error) {
	switch resource {
	case publicresource.Users:
		return layoutapi.ResourceUser, nil
	case publicresource.Roles:
		return layoutapi.ResourceRole, nil
	case publicresource.Organizations:
		return layoutapi.ResourceOrganization, nil
	default:
		return "", ErrContract
	}
}
