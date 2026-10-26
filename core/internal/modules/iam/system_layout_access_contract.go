package iam

import publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"

var (
	ErrSystemLayoutAccessAuthentication = publicresource.ErrLayoutAccessAuthentication
	ErrSystemLayoutAccessForbidden      = publicresource.ErrLayoutAccessForbidden
	ErrSystemLayoutAccessUnavailable    = publicresource.ErrLayoutAccessUnavailable
)

type SystemLayoutAccessRequest = publicresource.LayoutAccessRequest
type SystemLayoutAccessDecision = publicresource.LayoutAccessDecision

var NewSystemLayoutAccessDecision = publicresource.NewLayoutAccessDecision
