package iam

import publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"

type identitySystemResourcePort struct {
	application *identityHTTPApplication
}

func (application *identityHTTPApplication) SystemResources() publicresource.Port {
	if application == nil {
		return nil
	}
	return &identitySystemResourcePort{application: application}
}

var _ publicresource.Port = (*identitySystemResourcePort)(nil)
var _ publicresource.LayoutCatalogPort = (*identitySystemResourcePort)(nil)
